package pay

const (
	// ErrPayJSONDecodeFailed 解析json出错
	ErrPayJSONDecodeFailed = 1

	// ErrPayGetPfAndPfkeyFailed 获取pf和pfkey失败
	ErrPayGetPfAndPfkeyFailed = 2

	// ErrPayNewRequestFailed 生成向midas的请求失败
	ErrPayNewRequestFailed = 3

	// ErrPayReqMidasFailed 请求midas失败
	ErrPayReqMidasFailed = 4

	// ErrPayReadMidasRetFailed 读取midas返回数据失败
	ErrPayReadMidasRetFailed = 5

	// ErrPayBuyPropIDFailed 购买物品时物品id错误
	ErrPayBuyPropIDFailed = 5
)

// 米大师appid
const (
	AndroidAppid = "1450013440"
	IOSAppid     = "1450013441"
)
