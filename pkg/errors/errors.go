package errors

import "net/http"

type Error struct {
	Type    string
	Message string
	Status  int
}

// 4xx Client Errors
// 클라이언트에 오류가 있음
func BadRequest(msg string) Error {
	return Error{Type: "BadRequest", Message: msg, Status: http.StatusBadRequest} // 400: 요청이 잘못되었음
}

func Unauthorized(msg string) Error {
	return Error{Type: "Unauthorized", Message: msg, Status: http.StatusUnauthorized} // 401: 인증이 필요함
}

func PaymentRequired(msg string) Error {
	return Error{Type: "PaymentRequired", Message: msg, Status: http.StatusPaymentRequired} // 402: 결제가 필요함
}

func Forbidden(msg string) Error {
	return Error{Type: "Forbidden", Message: msg, Status: http.StatusForbidden} // 403: 요청이 거부됨
}

func NotFound(msg string) Error {
	return Error{Type: "NotFound", Message: msg, Status: http.StatusNotFound} // 404: 요청한 리소스가 없음
}

func MethodNotAllowed(msg string) Error {
	return Error{Type: "MethodNotAllowed", Message: msg, Status: http.StatusMethodNotAllowed} // 405: 요청된 메소드가 허용되지 않음
}

func NotAcceptable(msg string) Error {
	return Error{Type: "NotAcceptable", Message: msg, Status: http.StatusNotAcceptable} // 406: 요청된 리소스가 클라이언트가 허용하지 않음
}

func ProxyAuthRequired(msg string) Error {
	return Error{Type: "ProxyAuthRequired", Message: msg, Status: http.StatusProxyAuthRequired} // 407: 프록시 인증이 필요함
}

func RequestTimeout(msg string) Error {
	return Error{Type: "RequestTimeout", Message: msg, Status: http.StatusRequestTimeout} // 408: 요청 시간이 초과됨
}

func Conflict(msg string) Error {
	return Error{Type: "Conflict", Message: msg, Status: http.StatusConflict} // 409: 요청이 충돌함
}

func Gone(msg string) Error {
	return Error{Type: "Gone", Message: msg, Status: http.StatusGone} // 410: 요청한 리소스가 더 이상 사용되지 않음
}

func LengthRequired(msg string) Error {
	return Error{Type: "LengthRequired", Message: msg, Status: http.StatusLengthRequired} // 411: Content-Length 헤더가 필요함
}

func PreconditionFailed(msg string) Error {
	return Error{Type: "PreconditionFailed", Message: msg, Status: http.StatusPreconditionFailed} // 412: 요청 전제 조건이 실패함
}

func PayloadTooLarge(msg string) Error {
	return Error{Type: "PayloadTooLarge", Message: msg, Status: http.StatusRequestEntityTooLarge} // 413: 요청이 너무 큼
}

func URITooLong(msg string) Error {
	return Error{Type: "URITooLong", Message: msg, Status: http.StatusRequestURITooLong} // 414: URI가 너무 김
}

func UnsupportedMediaType(msg string) Error {
	return Error{Type: "UnsupportedMediaType", Message: msg, Status: http.StatusUnsupportedMediaType} // 415: 지원하지 않는 미디어 타입
}

func RangeNotSatisfiable(msg string) Error {
	return Error{Type: "RangeNotSatisfiable", Message: msg, Status: http.StatusRequestedRangeNotSatisfiable} // 416: 범위가 만족되지 않음
}

func ExpectationFailed(msg string) Error {
	return Error{Type: "ExpectationFailed", Message: msg, Status: http.StatusExpectationFailed} // 417: 요청이 실패함
}

func Teapot(msg string) Error {
	return Error{Type: "Teapot", Message: msg, Status: http.StatusTeapot} // 418: 나는 주전자입니다
}

func MisdirectedRequest(msg string) Error {
	return Error{Type: "MisdirectedRequest", Message: msg, Status: http.StatusMisdirectedRequest} // 421: 잘못된 요청
}

func UnprocessableEntity(msg string) Error {
	return Error{Type: "UnprocessableEntity", Message: msg, Status: http.StatusUnprocessableEntity} // 422: 처리할 수 없는 엔티티
}

func Locked(msg string) Error {
	return Error{Type: "Locked", Message: msg, Status: http.StatusLocked} // 423: 잠김
}

func FailedDependency(msg string) Error {
	return Error{Type: "FailedDependency", Message: msg, Status: http.StatusFailedDependency} // 424: 의존성 실패
}

func TooEarly(msg string) Error {
	return Error{Type: "TooEarly", Message: msg, Status: http.StatusTooEarly} // 425: 너무 이른 요청
}

func UpgradeRequired(msg string) Error {
	return Error{Type: "UpgradeRequired", Message: msg, Status: http.StatusUpgradeRequired} // 426: 업그레이드 필요
}

func PreconditionRequired(msg string) Error {
	return Error{Type: "PreconditionRequired", Message: msg, Status: http.StatusPreconditionRequired} // 428: 전제 조건 필요
}

func TooManyRequests(msg string) Error {
	return Error{Type: "TooManyRequests", Message: msg, Status: http.StatusTooManyRequests} // 429: 요청이 너무 많음
}

func RequestHeaderFieldsTooLarge(msg string) Error {
	return Error{Type: "RequestHeaderFieldsTooLarge", Message: msg, Status: http.StatusRequestHeaderFieldsTooLarge} // 431: 요청 헤더 필드가 너무 큼
}

func UnavailableForLegalReasons(msg string) Error {
	return Error{Type: "UnavailableForLegalReasons", Message: msg, Status: http.StatusUnavailableForLegalReasons} // 451: 법적 이유로 사용할 수 없음
}

// 5xx Server Errors
// 서버에 오류가 있음
func Internal(msg string) Error {
	return Error{Type: "Internal", Message: msg, Status: http.StatusInternalServerError} // 500: 서버에 오류가 있음
}

func NotImplemented(msg string) Error {
	return Error{Type: "NotImplemented", Message: msg, Status: http.StatusNotImplemented} // 501: 요청이 구현되지 않음
}

func BadGateway(msg string) Error {
	return Error{Type: "BadGateway", Message: msg, Status: http.StatusBadGateway} // 502: 게이트웨이가 잘못됨
}

func ServiceUnavailable(msg string) Error {
	return Error{Type: "ServiceUnavailable", Message: msg, Status: http.StatusServiceUnavailable} // 503: 서비스를 사용할 수 없음
}

func GatewayTimeout(msg string) Error {
	return Error{Type: "GatewayTimeout", Message: msg, Status: http.StatusGatewayTimeout} // 504: 게이트웨이 시간 초과
}

func HTTPVersionNotSupported(msg string) Error {
	return Error{Type: "HTTPVersionNotSupported", Message: msg, Status: http.StatusHTTPVersionNotSupported} // 505: HTTP 버전이 지원되지 않음
}

func VariantAlsoNegotiates(msg string) Error {
	return Error{Type: "VariantAlsoNegotiates", Message: msg, Status: http.StatusVariantAlsoNegotiates} // 506: 변형도 협상함
}

func InsufficientStorage(msg string) Error {
	return Error{Type: "InsufficientStorage", Message: msg, Status: http.StatusInsufficientStorage} // 507: 저장 공간이 부족함
}

func LoopDetected(msg string) Error {
	return Error{Type: "LoopDetected", Message: msg, Status: http.StatusLoopDetected} // 508: 루프가 감지됨
}

func NotExtended(msg string) Error {
	return Error{Type: "NotExtended", Message: msg, Status: http.StatusNotExtended} // 510: 확장되지 않음
}

func NetworkAuthenticationRequired(msg string) Error {
	return Error{Type: "NetworkAuthenticationRequired", Message: msg, Status: http.StatusNetworkAuthenticationRequired} // 511: 네트워크 인증이 필요함
}

// Custom error creator
func New(status int, errType string, msg string) Error {
	return Error{
		Status:  status,
		Type:    errType,
		Message: msg,
	}
}
