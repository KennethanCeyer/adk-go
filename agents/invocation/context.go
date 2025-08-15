package invocation

import (
	"context"
	"fmt"
)

type contextKey string

const uiSenderKey = contextKey("uiSender")

// UISender 는 연결된 UI 클라이언트로 메시지를 보내는 함수입니다.
type UISender func(messageType string, payload any)

// WithUISender 는 주어진 UISender를 포함하는 새로운 context를 반환합니다.
func WithUISender(ctx context.Context, sender UISender) context.Context {
	return context.WithValue(ctx, uiSenderKey, sender)
}

// GetUISender 는 context에서 UISender를 검색합니다.
func GetUISender(ctx context.Context) (UISender, bool) {
	sender, ok := ctx.Value(uiSenderKey).(UISender)
	return sender, ok
}

// SendInternalLog 는 context에 sender가 있을 경우 UI로 내부 로그 메시지를 보내는 헬퍼 함수입니다.
func SendInternalLog(ctx context.Context, format string, a ...any) {
	if sender, ok := GetUISender(ctx); ok {
		text := fmt.Sprintf(format, a...)
		sender("internal_log", map[string]string{"text": text})
	}
}
