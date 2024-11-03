package errors

import "net/http"

type CustomError interface {
	error
	GetType() string
	GetStatus() int
	GetMessage() string
}

type errorImpl struct {
	Type    string
	Message string
	Status  int
}

func (e errorImpl) Error() string {
	return e.Message
}

func (e errorImpl) GetType() string {
	return e.Type
}

func (e errorImpl) GetStatus() int {
	return e.Status
}

func (e errorImpl) GetMessage() string {
	return e.Message
}

// 4xx Client Errors
// 클라이언트에 오류가 있음
func BadRequest(msg string) errorImpl {
	return errorImpl{Type: "BadRequest", Message: msg, Status: http.StatusBadRequest} // 400: 요청이 잘못되었음
}

func Unauthorized(msg string) errorImpl {
	return errorImpl{Type: "Unauthorized", Message: msg, Status: http.StatusUnauthorized} // 401: 인증이 필요함
}

func PaymentRequired(msg string) errorImpl {
	return errorImpl{Type: "PaymentRequired", Message: msg, Status: http.StatusPaymentRequired} // 402: 결제가 필요함
}

func Forbidden(msg string) errorImpl {
	return errorImpl{Type: "Forbidden", Message: msg, Status: http.StatusForbidden} // 403: 요청이 거부됨
}

func NotFound(msg string) errorImpl {
	return errorImpl{Type: "NotFound", Message: msg, Status: http.StatusNotFound} // 404: 요청한 리소스가 없음
}

func MethodNotAllowed(msg string) errorImpl {
	return errorImpl{Type: "MethodNotAllowed", Message: msg, Status: http.StatusMethodNotAllowed} // 405: 요청된 메소드가 허용되지 않음
}

func NotAcceptable(msg string) errorImpl {
	return errorImpl{Type: "NotAcceptable", Message: msg, Status: http.StatusNotAcceptable} // 406: 요청된 리소스가 클라이언트가 허용하지 않음
}

func ProxyAuthRequired(msg string) errorImpl {
	return errorImpl{Type: "ProxyAuthRequired", Message: msg, Status: http.StatusProxyAuthRequired} // 407: 프록시 인증이 필요함
}

func RequestTimeout(msg string) errorImpl {
	return errorImpl{Type: "RequestTimeout", Message: msg, Status: http.StatusRequestTimeout} // 408: 요청 시간이 초과됨
}

func Conflict(msg string) errorImpl {
	return errorImpl{Type: "Conflict", Message: msg, Status: http.StatusConflict} // 409: 요청이 충돌함
}

func Gone(msg string) errorImpl {
	return errorImpl{Type: "Gone", Message: msg, Status: http.StatusGone} // 410: 요청한 리소스가 더 이상 사용되지 않음
}

func LengthRequired(msg string) errorImpl {
	return errorImpl{Type: "LengthRequired", Message: msg, Status: http.StatusLengthRequired} // 411: Content-Length 헤더가 필요함
}

func PreconditionFailed(msg string) errorImpl {
	return errorImpl{Type: "PreconditionFailed", Message: msg, Status: http.StatusPreconditionFailed} // 412: 요청 전제 조건이 실패함
}

func PayloadTooLarge(msg string) errorImpl {
	return errorImpl{Type: "PayloadTooLarge", Message: msg, Status: http.StatusRequestEntityTooLarge} // 413: 요청이 너무 큼
}

func URITooLong(msg string) errorImpl {
	return errorImpl{Type: "URITooLong", Message: msg, Status: http.StatusRequestURITooLong} // 414: URI가 너무 김
}

func UnsupportedMediaType(msg string) errorImpl {
	return errorImpl{Type: "UnsupportedMediaType", Message: msg, Status: http.StatusUnsupportedMediaType} // 415: 지원하지 않는 미디어 타입
}

func RangeNotSatisfiable(msg string) errorImpl {
	return errorImpl{Type: "RangeNotSatisfiable", Message: msg, Status: http.StatusRequestedRangeNotSatisfiable} // 416: 범위가 만족되지 않음
}

func ExpectationFailed(msg string) errorImpl {
	return errorImpl{Type: "ExpectationFailed", Message: msg, Status: http.StatusExpectationFailed} // 417: 요청이 실패함
}

func Teapot(msg string) errorImpl {
	return errorImpl{Type: "Teapot", Message: msg, Status: http.StatusTeapot} // 418: 나는 주전자입니다
}

func MisdirectedRequest(msg string) errorImpl {
	return errorImpl{Type: "MisdirectedRequest", Message: msg, Status: http.StatusMisdirectedRequest} // 421: 잘못된 요청
}

func UnprocessableEntity(msg string) errorImpl {
	return errorImpl{Type: "UnprocessableEntity", Message: msg, Status: http.StatusUnprocessableEntity} // 422: 처리할 수 없는 엔티티
}

func Locked(msg string) errorImpl {
	return errorImpl{Type: "Locked", Message: msg, Status: http.StatusLocked} // 423: 잠김
}

func FailedDependency(msg string) errorImpl {
	return errorImpl{Type: "FailedDependency", Message: msg, Status: http.StatusFailedDependency} // 424: 의존성 실패
}

func TooEarly(msg string) errorImpl {
	return errorImpl{Type: "TooEarly", Message: msg, Status: http.StatusTooEarly} // 425: 너무 이른 요청
}

func UpgradeRequired(msg string) errorImpl {
	return errorImpl{Type: "UpgradeRequired", Message: msg, Status: http.StatusUpgradeRequired} // 426: 업그레이드 필요
}

func PreconditionRequired(msg string) errorImpl {
	return errorImpl{Type: "PreconditionRequired", Message: msg, Status: http.StatusPreconditionRequired} // 428: 전제 조건 필요
}

func TooManyRequests(msg string) errorImpl {
	return errorImpl{Type: "TooManyRequests", Message: msg, Status: http.StatusTooManyRequests} // 429: 요청이 너무 많음
}

func RequestHeaderFieldsTooLarge(msg string) errorImpl {
	return errorImpl{Type: "RequestHeaderFieldsTooLarge", Message: msg, Status: http.StatusRequestHeaderFieldsTooLarge} // 431: 요청 헤더 필드가 너무 큼
}

func UnavailableForLegalReasons(msg string) errorImpl {
	return errorImpl{Type: "UnavailableForLegalReasons", Message: msg, Status: http.StatusUnavailableForLegalReasons} // 451: 법적 이유로 사용할 수 없음
}

// 5xx Server Errors
// 서버에 오류가 있음
func Internal(msg string) errorImpl {
	return errorImpl{Type: "Internal", Message: msg, Status: http.StatusInternalServerError} // 500: 서버에 오류가 있음
}

func NotImplemented(msg string) errorImpl {
	return errorImpl{Type: "NotImplemented", Message: msg, Status: http.StatusNotImplemented} // 501: 요청이 구현되지 않음
}

func BadGateway(msg string) errorImpl {
	return errorImpl{Type: "BadGateway", Message: msg, Status: http.StatusBadGateway} // 502: 게이트웨이가 잘못됨
}

func ServiceUnavailable(msg string) errorImpl {
	return errorImpl{Type: "ServiceUnavailable", Message: msg, Status: http.StatusServiceUnavailable} // 503: 서비스를 사용할 수 없음
}

func GatewayTimeout(msg string) errorImpl {
	return errorImpl{Type: "GatewayTimeout", Message: msg, Status: http.StatusGatewayTimeout} // 504: 게이트웨이 시간 초과
}

func HTTPVersionNotSupported(msg string) errorImpl {
	return errorImpl{Type: "HTTPVersionNotSupported", Message: msg, Status: http.StatusHTTPVersionNotSupported} // 505: HTTP 버전이 지원되지 않음
}

func VariantAlsoNegotiates(msg string) errorImpl {
	return errorImpl{Type: "VariantAlsoNegotiates", Message: msg, Status: http.StatusVariantAlsoNegotiates} // 506: 변형도 협상함
}

func InsufficientStorage(msg string) errorImpl {
	return errorImpl{Type: "InsufficientStorage", Message: msg, Status: http.StatusInsufficientStorage} // 507: 저장 공간이 부족함
}

func LoopDetected(msg string) errorImpl {
	return errorImpl{Type: "LoopDetected", Message: msg, Status: http.StatusLoopDetected} // 508: 루프가 감지됨
}

func NotExtended(msg string) errorImpl {
	return errorImpl{Type: "NotExtended", Message: msg, Status: http.StatusNotExtended} // 510: 확장되지 않음
}

func NetworkAuthenticationRequired(msg string) errorImpl {
	return errorImpl{Type: "NetworkAuthenticationRequired", Message: msg, Status: http.StatusNetworkAuthenticationRequired} // 511: 네트워크 인증이 필요함
}

// Custom error creator
func New(status int, errType string, msg string) errorImpl {
	return errorImpl{
		Status:  status,
		Type:    errType,
		Message: msg,
	}
}
