package timer

import (
	"awesome/model"
	"awesome/lib/socket"
	"strconv"
	"time"
	"strings"
	"github.com/golang/glog"
)

const separator  = ":"

func Run() {
	DoSomething = writeToRoom
	var inteval = time.Second
	glog.Infof("时间轮启动: %s s",inteval)

	for {
		go step()        //TODO 保证此方法在1秒内执行完成
		<- time.After(inteval)
	}
}

func writeToRoom(i interface{}) {
	if key,ok := i.(string);!ok {
		glog.Errorf("%v is  no int",i)
	}else {
		eventStruct := strings.Split(key,separator)
		inviteCode,err := strconv.ParseFloat(eventStruct[0],64)
		if err != nil {
			glog.Errorln("timer not parse inviteCode ",key)
		}else {
			var found = false
			model.RoomRange(func(k,v interface{})bool{
				if k == int(inviteCode) {
					if room,ok := v.(*model.Room);ok {
						// 发送定时任务的类型
							room.WriteSysMsg(BuildTimeOutMsg(key))
					}else {
						glog.Errorln("not model.Room ",k,key)
					}
					found = true
					return false
				}
				return true
			})

			if  !found {
				if len(eventStruct) > 1 {
					glog.Errorf("准备销毁")
					model.NilRoomHelper.WriteSysMsg(BuildTimeOutMsg(key))
				}else {
					glog.Errorln("not find model.Room ", inviteCode)
				}
			}
		}
	}
}

func BuildTimeOutMsg(key string) *socket.PackHead {
	body := []byte(key)
	var msg = &socket.PackHead{Sid: 0, Body:body,Length:uint32(len(body)), Cmd: uint32(20)}
	socket.SerializePackHead(msg)
	return msg
}