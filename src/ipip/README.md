[17mon](http://www.ipip.net/) IP location data for Golang
===

## 特性
* 高效的查找算法，查询性能100w/s
* 支持build出的bin文件包含原始数据

## 定义
type LocationInfo struct {
	Country  string
	Province string
	City     string
	Isp      string
}

## 使用
	import （
		"fmt"
		"ipip"
	）

	func init() {
		if err := ipip.Init("your data file"); err != nil {
			panic(err)
		}
	}

	func main() {
		loc, err := ipip.Find("116.228.111.18")
		if err != nil {
			fmt.Println("err:", err)
			return
		}
		fmt.Println(loc)
	}


