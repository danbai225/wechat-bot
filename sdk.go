package wechat_bot

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/lxzan/gws"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const HeartBeat = 5005
const RecvTxtMsg = 1
const RecvPicMsg = 3
const UserList = 5000
const TxtMsg = 555
const PicMsg = 500
const AtMsg = 550
const PersonalDetail = 6550
const AttatchFile = 5003
const RecvFileMsg = 49
const PersonalInfo = 6500
const ChatroomMemberNick = 5020

func gid() string {
	u, _ := uuid.NewV4()
	return u.String()
}

func NewClient(ws, qrHttp string) (*Client, error) {
	c := new(Client)
	c.handler = handler{
		Client: c,
	}
	socket, _, err := gws.NewClient(&c.handler, &gws.ClientOption{
		Addr: ws,
	})
	if err != nil {
		return nil, err
	}
	c.socket = socket
	c.addr = ws
	c.qrAddr = qrHttp
	c.lastTime = time.Now().Unix()
	c.contactListChan = make(chan []*Contact, 0)
	c.infoChan = make(chan *Info, 0)
	go socket.ReadLoop()
	c.ticker = time.NewTicker(time.Second * 5)
	go c.healthExamination()
	return c, err
}

type Client struct {
	handler         handler
	socket          *gws.Conn
	addr            string
	qrAddr          string
	onMsg           func(msg []byte, Type int, reply *Reply) //type 1是文本 2是图片 3是文件
	contactListChan chan []*Contact
	infoChan        chan *Info
	lastTime        int64 //最后心跳时间
	lock            sync.Mutex
	ticker          *time.Ticker
}

func (c *Client) ShutDown() error {
	c.ticker.Stop()
	if c.socket == nil {
		return nil
	}
	return c.socket.NetConn().Close()
}

// RCon 重连
func (c *Client) RCon() error {
	if c.socket != nil {
		_ = c.socket.NetConn().Close()
	}
	socket, _, err := gws.NewClient(&c.handler, &gws.ClientOption{
		Addr: c.addr,
	})
	if err != nil {
		return err
	}
	c.socket = socket
	c.lastTime = time.Now().Unix()
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "重连成功")
	go socket.ReadLoop()
	return err
}
func (c *Client) healthExamination() {
	for {
		t := <-c.ticker.C
		if t.Unix()-c.lastTime > 100 || c.socket == nil {
			fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "重连...")
			_ = c.RCon()
		}
	}
}

type handler struct {
	Client *Client
	gws.BuiltinEventHandler
}

// Reply 快速回复
type Reply struct {
	c     *Client
	wId   string
	wxId1 string
	id    string
	nick  string
}

func (r *Reply) GetMsgID() string {
	return r.id
}
func (r *Reply) Msg(content string) error {
	return r.c.SendTxt(content, r.wId)
}
func (r *Reply) GetNick() string {
	if r.nick == "" {
		r.nick, _ = r.c.GetNickFormRoom(r.wxId1, r.wId)
	}
	return r.nick
}
func (r *Reply) AtMsg(content string) error {
	return r.c.SendAtMsg(content, r.wxId1, r.wId, r.GetNick())
}
func (r *Reply) PrivateChat(content string) error {
	return r.c.SendTxt(content, r.wxId1)
}
func (r *Reply) PicMsg(path string) error {
	return r.c.SendPicMsg(path, r.wId)
}
func (r *Reply) PrivatePicMsg(path string) error {
	return r.c.SendPicMsg(path, r.wxId1)
}
func (r *Reply) File(path string) error {
	return r.c.SendFile(path, r.wId)
}
func (r *Reply) PrivateFile(path string) error {
	return r.c.SendFile(path, r.wxId1)
}
func (r *Reply) Bytes2Path(data []byte) (string, error) {
	sum := md5.Sum(data)
	md5str := fmt.Sprintf("%x", sum)
	path := fmt.Sprintf("%s%c%s", os.TempDir(), os.PathSeparator, md5str)
	err := os.WriteFile(path, data, 0666)
	return path, err
}

// IsSendByFriend 来自私聊
func (r *Reply) IsSendByFriend() bool {
	if strings.Contains(r.wId, "@chatroom") {
		return false
	} else {
		return true
	}
}
func (r *Reply) IsSendByGroup() bool {
	if strings.Contains(r.wId, "@chatroom") {
		return true
	} else {
		return false
	}
}

// GetWxID 如果是群消息，返回群ID，如果是私聊消息，返回用户ID
func (r *Reply) GetWxID() string {
	return r.wId
}

// GetPrivateWxID 返回用户ID
func (r *Reply) GetPrivateWxID() string {
	if r.IsSendByGroup() {
		return r.wxId1
	} else {
		return r.wId
	}
}

type msg struct {
	Id       string      `json:"id"`
	Type     int         `json:"type"`
	Wxid     interface{} `json:"wxid"`
	Roomid   interface{} `json:"roomid"`
	Content  interface{} `json:"content"`
	Nickname interface{} `json:"nickname"`
	Ext      interface{} `json:"ext"`
}
type ImgMsg struct {
	Content string `json:"content"`
	Detail  string `json:"detail"`
	Id1     string `json:"id1"`
	Id2     string `json:"id2"`
	Thumb   string `json:"thumb"`
}

func ParsePictureMessage(msg []byte) *ImgMsg {
	im := &ImgMsg{}
	_ = json.Unmarshal(msg, im)
	return im
}

type Dic struct {
	name       string //名称
	firstIndex uint8  //第一个字节
	lastIndex  uint8  //第二个字节
}

var dicList = []Dic{{".jpg", 0xff, 0xd8},
	{".png", 0x89, 0x50},
	{".gif", 0x47, 0x49},
	{"error", 0x00, 0x00}}

func parseData(data []byte) {
	var addCode uint8
	for _, dic := range dicList {
		addCode = data[0] ^ dic.firstIndex
		if data[1]^addCode == dic.lastIndex {
			break
		}
	}
	//对字节切片每个字节异或
	for i, v := range data {
		data[i] = v ^ addCode
	}
}

func (im *ImgMsg) GetData(client *Client) ([]byte, error) {
	time.Sleep(time.Second)
	file, err := client.getFile(strings.ReplaceAll(im.Detail, "\\", "/"))
	if err != nil {
		return nil, err
	}
	parseData(file)
	return file, nil
}

// SetOnWXmsg type 1是文本 2是图片 3是文件
func (c *Client) SetOnWXmsg(onMsg func(msg []byte, Type int, reply *Reply)) {
	c.onMsg = onMsg
}
func (c *Client) LastHeartbeatTime() int64 {
	return c.lastTime
}
func (c *handler) OnMessage(socket *gws.Conn, message *gws.Message) {
	//fmt.Printf("recv: %s\n", message.Data.String())
	m := &rMsg{}
	_ = json.Unmarshal(message.Data.Bytes(), m)
	switch m.Type {
	case UserList:
		contacts := make([]*Contact, 0)
		marshal, _ := json.Marshal(m.Content)
		_ = json.Unmarshal(marshal, &contacts)
		c.Client.contactListChan <- contacts
	case HeartBeat:
		_ = c.Client.send(&msg{Id: gid(), Type: PicMsg, Wxid: "null", Roomid: "null", Content: "null", Nickname: "null", Ext: "null"})
		c.Client.lastTime = time.Now().Unix()
	case RecvTxtMsg:
		if c.Client.onMsg != nil {
			go c.Client.onMsg([]byte(m.Content.(string)), 1, &Reply{id: m.Id, c: c.Client, wId: m.WxID, wxId1: m.ID1})
		}
	case RecvPicMsg:
		if c.Client.onMsg != nil {
			marshal, _ := json.Marshal(m.Content)
			go c.Client.onMsg(marshal, 2, &Reply{id: m.Id, c: c.Client, wId: m.WxID})
		}
	case RecvFileMsg:
		if c.Client.onMsg != nil {
			marshal, _ := json.Marshal(m.Content)
			go c.Client.onMsg(marshal, 3, &Reply{id: m.Id, c: c.Client, wId: m.WxID})
		}
	case PersonalDetail:
		marshal, _ := json.Marshal(m.Content)
		info := &Info{}
		_ = json.Unmarshal(marshal, &info)
		c.Client.infoChan <- info
	case PersonalInfo:
		info := &Info{}
		_ = json.Unmarshal([]byte(m.Content.(string)), &info)
		c.Client.infoChan <- info

	case ChatroomMemberNick:
		info := &Info{}
		_ = json.Unmarshal([]byte(m.Content.(string)), &info)
		c.Client.infoChan <- info
	}
}
func (c *handler) OnError(socket *gws.Conn, err error) {
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "OnError", err, socket.RemoteAddr().String())
	c.Client.socket = nil
}
func (c *handler) OnClose(socket *gws.Conn, err error) {
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "close")
	c.Client.socket = nil
}

func (c *Client) send(msg *msg) error {
	if c.socket == nil {
		return errors.New("socket is nil")
	}
	marshal, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.socket.WriteString(string(marshal))
}

/*QR
* 获取扫描登录二维码
 */
func (c *Client) QR() ([]byte, error) {
	resp, _ := http.Get(c.qrAddr + "/qr")
	return io.ReadAll(resp.Body)
}
func (c *Client) upFile(data []byte, filename string) error {
	// 创建multipart writer
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// 创建文件数据部分
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return err
	}
	_, err = part.Write(data)
	if err != nil {
		return err
	}
	// 写入multipart部分数据（文件名等）
	if err = writer.WriteField("filename", filename); err != nil {
		return err
	}
	if err = writer.Close(); err != nil {
		return err
	}
	// 创建HTTP请求并设置multipart数据
	req, err := http.NewRequest("POST", c.qrAddr+"/file", body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送HTTP请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
func (c *Client) getFile(path string) ([]byte, error) {
	resp, err := http.Get(c.qrAddr + "/download?path=" + path)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(resp.Body)
}

/*SendTxt
*发送文本消息
*content:消息内容
*to:接收者的wxid 个人消息为个人wxid 群消息为群wxid
 */
func (c *Client) SendTxt(content string, to string) error {
	return c.send(&msg{
		Id:       "",
		Type:     TxtMsg,
		Wxid:     to,
		Roomid:   "null",
		Content:  content,
		Nickname: "null",
		Ext:      "null",
	})
}

/*
SendAtMsg
*发送@消息
*content:消息内容
*atWXid:被@的人的wxid
*to:群wxid
*nickname:被@的人的昵称
*/
func (c *Client) SendAtMsg(content string, atWXid, to, nickname string) error {
	return c.send(&msg{
		Id:       gid(),
		Type:     AtMsg,
		Wxid:     atWXid,
		Roomid:   to,
		Content:  content,
		Nickname: nickname,
		Ext:      "null",
	})
}

/*
SendPicMsg
*发送图片信息
*content:消息内容
*atWXid:被@的人的wxid
*to:群wxid
*nickname:被@的人的昵称
*/
func (c *Client) SendPicMsg(path, wxid string) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	name := filepath.Base(path)
	err = c.upFile(file, name)
	if err != nil {
		return err
	}
	return c.send(&msg{
		Id:       gid(),
		Type:     PicMsg,
		Wxid:     wxid,
		Roomid:   "null",
		Content:  "c:\\data\\" + name,
		Nickname: "null",
		Ext:      "null",
	})
}

/*
SendFile
*发送文件
*content:消息内容
*atWXid:被@的人的wxid
*to:群wxid
*nickname:被@的人的昵称
*/
func (c *Client) SendFile(path, wxid string) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	name := filepath.Base(path)
	err = c.upFile(file, name)
	if err != nil {
		return err
	}
	return c.send(&msg{
		Id:       gid(),
		Type:     AttatchFile,
		Wxid:     wxid,
		Roomid:   "null",
		Content:  "c:\\data\\" + name,
		Nickname: "null",
		Ext:      "null",
	})
}
func (c *Client) GetContactList() ([]*Contact, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	err := c.send(&msg{
		Id:       gid(),
		Type:     UserList,
		Wxid:     "null",
		Roomid:   "null",
		Content:  "null",
		Nickname: "null",
		Ext:      "null",
	})
	if err != nil {
		return nil, err
	}

	return <-c.contactListChan, nil
}
func (c *Client) GetPersonalDetail(wxid string) (*Info, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	err := c.send(&msg{
		Id:       gid(),
		Type:     PersonalDetail,
		Wxid:     wxid,
		Roomid:   "null",
		Content:  "op:personal detail",
		Nickname: "null",
		Ext:      "null",
	})
	if err != nil {
		return nil, err
	}

	return <-c.infoChan, nil
}
func (c *Client) GetPersonal() (*Info, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	err := c.send(&msg{
		Id:       gid(),
		Type:     PersonalInfo,
		Wxid:     "ROOT",
		Roomid:   "null",
		Content:  "op:personal info",
		Nickname: "null",
		Ext:      "null",
	})
	if err != nil {
		return nil, err
	}

	return <-c.infoChan, nil
}
func (c *Client) GetNickFormRoom(wxid, roomid string) (string, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	err := c.send(&msg{
		Id:       gid(),
		Type:     ChatroomMemberNick,
		Wxid:     wxid,
		Roomid:   roomid,
		Content:  "null",
		Nickname: "null",
		Ext:      "null",
	})
	if err != nil {
		return "", err
	}
	i := <-c.infoChan
	return i.Nick, nil
}

type rMsg struct {
	Content  interface{} `json:"content"`
	Id       string      `json:"id"`
	Receiver string      `json:"receiver"`
	WxID     string      `json:"wxid"`
	ID1      string      `json:"id1"`
	Sender   string      `json:"sender"`
	Srvid    int         `json:"srvid"`
	Status   string      `json:"status"`
	Time     string      `json:"time"`
	Type     int         `json:"type"`
}

// Contact 联系人
type Contact struct {
	Headimg string `json:"headimg"`
	Name    string `json:"name"`
	Node    int    `json:"node"`
	Remarks string `json:"remarks"`
	Wxcode  string `json:"wxcode"`
	Wxid    string `json:"wxid"`
}

// Info 该结构体多个返回结果在公用不保证所有字段都有，自行判断
type Info struct {
	BigHeadimg    string `json:"big_headimg"`
	Cover         string `json:"cover"`
	LittleHeadimg string `json:"little_headimg"`
	Signature     string `json:"signature"`
	WxCode        string `json:"wx_code"`
	WxHeadImage   string `json:"wx_head_image"`
	WxId          string `json:"wx_id"`
	WxName        string `json:"wx_name"`
	Nick          string `json:"nick"`
}
