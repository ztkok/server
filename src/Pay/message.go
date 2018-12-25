package pay

// QueryBalace 客户端查询余额请求
type QueryBalaceByClient struct {
	SessionID      string `json:"SessionID"`      // 用户账户类型
	SessionType    string `json:"SessionType"`    // session类型
	Openid         string `json:"Openid"`         // openid
	Openkey        string `json:"Openkey"`        // openkey
	PayToken       string `json:"PayToken"`       // pay_token
	AccessToken    string `json:"AccessToken"`    // 登录态(qq使用paytoken，微信使用accesstoken)
	Platform       string `json:"Platform"`       //平台标识(一般情况下：qq对应值为desktop_m_qq，wx对应值为desktop_m_wx，guestwx对应值为desktop_m_guest)
	RegChannel     string `json:"RegChannel"`     //注册渠道
	Os             string `json:"Os"`             //系统(安卓对应android，ios对应iap)
	Installchannel string `json:"Installchannel"` //安装渠道
	Offerid        string `json:"Offerid"`        //支付的appid
	Pf             string `json:"Pf"`             //Pf
	PfKey          string `json:"PfKey"`          //PfKey
}

// MidasQueryBalanceTssList midas查询虚拟币余额结果月卡信息
type MidasQueryBalanceTssList struct {
	Innerproductid         string `json:"inner_productid"`        // 用户开通的订阅物品id
	Begintime              string `json:"begintime"`              // 用户订阅的开始时间
	Endtime                string `json:"endtime"`                // 用户订阅的结束时间
	Paychan                string `json:"paychan"`                // 用户订阅该物品id最后一次的支付渠道
	Paysubchan             uint32 `json:"paysubchan"`             //用户订阅该物品 id 最后一次的支付子渠道 id
	Autopaychan            string `json:"autopaychan"`            //预留扩展字段，目前没有使用
	Autopaysubchan         uint32 `json:"autopaysubchan"`         //预留扩展字段，目前没有使用
	Grandtotal_opendays    uint32 `json:"grandtotal_opendays"`    //用户订阅累计开通天数
	Grandtotal_presentdays uint32 `json:"grandtotal_presentdays"` //用户订阅累计赠送天数
	First_buy_time         string `json:"first_buy_time"`         //首充开通时间
	Extend                 string `json:"extend"`                 //预留扩展字段，目前没有使用
}

// MidasQueryBalanceResult midas查询虚拟币余额结果
type MidasQueryBalanceResult struct {
	Ret         int                        `json:"ret"`         // 返回码
	Balance     uint32                     `json:"balance"`     // 虚拟币个数（包含了赠送虚拟币）
	Gen_balance uint32                     `json:"gen_balance"` // 赠送虚拟币个数
	First_save  uint32                     `json:"first_save"`  // 是否满足首次充值， 1： 满足， 0： 不满足
	Save_amt    uint32                     `json:"save_amt"`    // 累计充值金额的虚拟币数量
	Gen_expire  uint32                     `json:"gen_expire"`  // 该字段已作废
	Tss_list    []MidasQueryBalanceTssList `json:"tss_list"`    // 月卡信息字段, 如果没有月卡该字段值为空
	Save_sum    uint32                     `json:"save_sum"`    // 历史总游戏币金额
	Cost_sum    uint32                     `json:"cost_sum"`    // 历史总消费游戏币金额
	Present_sum uint32                     `json:"present_sum"` // 历史累计收到赠送金额
}

// DeductVirtualCoin 客户端扣除虚拟货币请求
type DeductVirtualCoin struct {
	SessionID      string `json:"SessionID"`      // 用户账户类型
	SessionType    string `json:"SessionType"`    // session类型
	Openid         string `json:"Openid"`         // openid
	Openkey        string `json:"Openkey"`        // openkey
	PayToken       string `json:"PayToken"`       // pay_token
	AccessToken    string `json:"AccessToken"`    // 登录态(qq使用paytoken，微信使用accesstoken)
	Platform       string `json:"Platform"`       //平台标识(一般情况下：qq对应值为desktop_m_qq，wx对应值为desktop_m_wx，guestwx对应值为desktop_m_guest)
	RegChannel     string `json:"RegChannel"`     //注册渠道
	Os             string `json:"Os"`             //系统(安卓对应android，ios对应iap)
	Installchannel string `json:"Installchannel"` //安装渠道
	Offerid        string `json:"Offerid"`        //支付的appid
	Pf             string `json:"Pf"`             //Pf
	PfKey          string `json:"PfKey"`          //PfKey
}

// MidasDeductVirtualCoinResult midas扣除虚拟货币结果
type MidasDeductVirtualCoinResult struct {
	Ret          int    // `json:"ret"`返回码
	Billno       string // `json:"billno"`预扣流水号
	Balance      uint32 // `json:"balance"`预扣后的余额
	Used_gen_amt uint32 // `json:"used_gen_amt"`支付使用赠送币金额
}

// GetpfAndpfkeyData 获取pf和pfkey需要的数据
type GetpfAndpfkeyData struct {
	Appid          string `json:"appid"`          // 游戏的唯一标识
	Openid         string `json:"openid"`         // 用户的唯一标识
	AccessToken    string `json:"accessToken"`    // openkey
	Platform       string `json:"platform"`       // pay_token
	RegChannel     string `json:"regChannel"`     // 登录态(qq使用paytoken，微信使用accesstoken)
	Os             string `json:"os"`             //系统(安卓对应android，ios对应iap)
	Installchannel string `json:"installchannel"` //安装渠道
	Offerid        string `json:"offerid"`        //支付的appid
}

// GetpfAndpfkeyRet 获取pf和pfkey返回信息
type GetpfAndpfkeyRet struct {
	Ret   int    `json:"ret"`   // 返回码 0：正确，其它：失败
	Msg   string `json:"msg"`   // ret非0，则表示“错误码，错误提示”，详细注释参见错误码描述
	Pf    string `json:"pf"`    // 对应的pf值
	PfKey string `json:"pfKey"` // 对应的pfKey值
}

type PresentMsg struct {
	SessionID      string `json:"SessionID"`      // 用户账户类型
	SessionType    string `json:"SessionType"`    // session类型
	Openid         string `json:"Openid"`         // openid
	Openkey        string `json:"Openkey"`        // openkey
	PayToken       string `json:"PayToken"`       // pay_token
	AccessToken    string `json:"AccessToken"`    // 登录态(qq使用paytoken，微信使用accesstoken)
	Platform       string `json:"Platform"`       //平台标识(一般情况下：qq对应值为desktop_m_qq，wx对应值为desktop_m_wx，guestwx对应值为desktop_m_guest)
	RegChannel     string `json:"RegChannel"`     //注册渠道
	Os             string `json:"Os"`             //系统(安卓对应android，ios对应iap)
	Installchannel string `json:"Installchannel"` //安装渠道
	Offerid        string `json:"Offerid"`        //支付的appid
	Pf             string `json:"Pf"`             //Pf
	PfKey          string `json:"PfKey"`          //PfKey

	// Discountid   string `json:"discountid"`
	// Giftid       string `json:"giftid"`
	// Presenttimes string `json:"presenttimes"`
}

type PresentRet struct {
	Ret     int    `json:"ret"` // 返回码 0：正确，其它：失败
	Billno  string // `json:"billno"`预扣流水号
	Balance uint32 `json:"balance"` // 预扣后的余额
	Msg     string `json:"msg"`     // ret非0，则表示“错误码，错误提示”，详细注释参见错误码描述
}

type CancelPayMsg struct {
	SessionID      string `json:"SessionID"`      // 用户账户类型
	SessionType    string `json:"SessionType"`    // session类型
	Openid         string `json:"Openid"`         // openid
	Openkey        string `json:"Openkey"`        // openkey
	PayToken       string `json:"PayToken"`       // pay_token
	AccessToken    string `json:"AccessToken"`    // 登录态(qq使用paytoken，微信使用accesstoken)
	Platform       string `json:"Platform"`       //平台标识(一般情况下：qq对应值为desktop_m_qq，wx对应值为desktop_m_wx，guestwx对应值为desktop_m_guest)
	RegChannel     string `json:"RegChannel"`     //注册渠道
	Os             string `json:"Os"`             //系统(安卓对应android，ios对应iap)
	Installchannel string `json:"Installchannel"` //安装渠道
	Offerid        string `json:"Offerid"`        //支付的appid
	Pf             string `json:"Pf"`             //Pf
	PfKey          string `json:"PfKey"`          //PfKey
}

type CancelPayRet struct {
	Ret     int    `json:"ret"` // 返回码 0：正确，其它：失败
	Billno  string // `json:"billno"`预扣流水号
	Balance uint32 `json:"balance"` // 预扣后的余额
	Msg     string `json:"msg"`     // ret非0，则表示“错误码，错误提示”，详细注释参见错误码描述
}
