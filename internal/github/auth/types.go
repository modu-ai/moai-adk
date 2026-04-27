package auth

import "context"

// AuthStatus는 현재 인증 상태를 나타냅니다.
type AuthStatus struct {
	Installed     bool   // CLI 또는 툴이 설치됨
	Authenticated bool   // 토큰/키가 구성됨
	SecretName    string // GitHub 시크릿 이름
	Message       string // 상태 메시지
}

// SecretSetter는 GitHub 시크릿 설정 인터페이스입니다.
type SecretSetter interface {
	SetSecret(ctx context.Context, repo, name, value string) error
}

// AuthHandler는 LLM 제공업체 인증 핸들러 인터페이스입니다.
type AuthHandler interface {
	Setup(ctx context.Context, repo string) error
}
