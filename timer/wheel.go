package timer

import (
	"sync"
)

const wheel_cnt uint8 = 5                                                                   //时间轮数量5个
var element_cnt_per_wheel = [wheel_cnt]uint32{256, 64, 64, 64, 64}                          //每个时间轮的槽(元素)数量。在 256+64+64+64+64 = 512 个槽中，表示的范围为 2^32
var right_shift_per_wheel = [wheel_cnt]uint32{8, 6, 6, 6, 6}                                //当指针指向当前时间轮最后一位数，再走一位就需要向上进位。每个时间轮进位的时候，使用右移的方式，最快实现进位。这里是每个轮的进位二进制位数
var base_per_wheel = [wheel_cnt]uint32{1, 256, 256 * 64, 256 * 64 * 64, 256 * 64 * 64 * 64} //记录每个时间轮指针当前指向的位置
var rwmutex sync.RWMutex
var newest [wheel_cnt]uint32                           //每个时间轮当前指针所指向的位置
var timewheels [5][]*Node                              //定义5个时间轮
var TimerMap map[string]*Node = make(map[string]*Node) //保存待执行的计时器，方便按链表节点指针地址直接删除定时器

type Timer struct {
	Key  		string            	//定时器Key   todo 此key设置为整形
	Inteval     uint32            	//时间间隔，即以插入该定时器的时间为起点，Inteval秒之后执行回调函数DoSomething()。例如进程插入该定时器的时间是2015-04-05 10:23:00，Inteval=5，则执行DoSomething()的时间就是2015-04-05 10:23:05。
}

func SetTimer(inviteCode string, inteval uint32) {
	if inteval <= 0 {
		return
	}
	var bucket_no uint8 = 0
	var offset uint32 = inteval
	var left uint32 = inteval
	for offset >= element_cnt_per_wheel[bucket_no] { //偏移量大于当前时间轮容量，则需要向高位进位
		offset >>= right_shift_per_wheel[bucket_no] //计算高位的值。偏移量除以低位的进制。比如低位当前是256，则右移8个二进制位，就是除以256，得到的结果是高位的值。
		var tmp uint32 = 1
		if bucket_no == 0 {
			tmp = 0
		}
		left -= base_per_wheel[bucket_no] * (element_cnt_per_wheel[bucket_no] - newest[bucket_no] - tmp)
		bucket_no++
	}
	if offset < 1 {
		return
	}
	if inteval < base_per_wheel[bucket_no]*offset {
		return
	}
	left -= base_per_wheel[bucket_no] * (offset - 1)
	pos := (newest[bucket_no] + offset) % element_cnt_per_wheel[bucket_no] //通过类似hash的方式，找到在时间轮上的插入位置

	var node Node
	node.SetData(Timer{inviteCode, left})


	rwmutex.Lock()
	Delete(TimerMap[inviteCode])
	TimerMap[inviteCode] = timewheels[bucket_no][pos].InsertHead(node) //插入定时器
	rwmutex.Unlock()
	//fmt.Println("pos ", bucket_no, pos, tmp)
}


var DoSomething = func(interface{}){}
func step() {
	//var dolist list.List
	{
		rwmutex.RLock()
		//遍历所有桶
		var bucket_no uint8 = 0
		for bucket_no = 0; bucket_no < wheel_cnt; bucket_no++ {
			newest[bucket_no] = (newest[bucket_no] + 1) % element_cnt_per_wheel[bucket_no] //当前指针递增1
			//fmt.Println(newest)
			var head *Node = timewheels[bucket_no][newest[bucket_no]] //返回当前指针指向的槽位置的表头
			var firstElement *Node = head.Next()
			for firstElement != nil { //链表不为空
				if val, ok := firstElement.Data().(Timer); ok { //如果element里面确实存储了Timer类型的数值，那么ok返回true，否则返回false。
					go DoSomething(val.Key)

					Delete(firstElement) //删除定时器
				}
				firstElement = head.Next() //重新定位到链表第一个元素头
			}
			if 0 != newest[bucket_no] { //指针不是0，还未转回到原点，跳出。如果回到原点，则说明转完了一圈，需要向高位进位1，则继续循环入高位步进一步。
				break
			}
		}
		rwmutex.RUnlock()
	}
}

func init() { //初始化
	var bucket_no uint8 = 0
	for bucket_no = 0; bucket_no < wheel_cnt; bucket_no++ {
		var i uint32 = 0
		for ; i < element_cnt_per_wheel[bucket_no]; i++ {
			timewheels[bucket_no] = append(timewheels[bucket_no], new(Node))
		}
	}
}
