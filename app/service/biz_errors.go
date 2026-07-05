package service

import errContract "bit-labs.cn/owl/contract/errors"

const (
	CodeAppNotFound      = "METRICS_APP_NOT_FOUND"
	CodeAppDisabled      = "METRICS_APP_DISABLED"
	CodeAppKeyDuplicate  = "METRICS_APP_KEY_DUPLICATE"
	CodeOriginForbidden  = "METRICS_ORIGIN_FORBIDDEN"
	CodeSignatureInvalid = "METRICS_SIGNATURE_INVALID"
	CodeCollectInvalid   = "METRICS_COLLECT_INVALID"
)

func AppNotFound() *errContract.BizError {
	return errContract.NewBizError(CodeAppNotFound, "指标应用不存在")
}

func AppDisabled() *errContract.BizError {
	return errContract.NewBizError(CodeAppDisabled, "指标应用已禁用")
}

func AppKeyDuplicate() *errContract.BizError {
	return errContract.NewBizError(CodeAppKeyDuplicate, "AppKey 已存在")
}

func OriginForbidden() *errContract.BizError {
	return errContract.NewBizError(CodeOriginForbidden, "来源不在白名单内")
}

func SignatureInvalid() *errContract.BizError {
	return errContract.NewBizError(CodeSignatureInvalid, "签名无效")
}

func CollectInvalid(msg string) *errContract.BizError {
	if msg == "" {
		msg = "采集参数无效"
	}
	return errContract.NewBizError(CodeCollectInvalid, msg)
}
