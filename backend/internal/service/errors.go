package service

import "errors"

var (
	// ErrInvalidCredentials 表示用户名或密码错误。
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrUnauthorized 表示缺乏身份凭证。
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden 表示权限不足。
	ErrForbidden = errors.New("forbidden")
	// ErrNotFound 表示资源不存在。
	ErrNotFound = errors.New("not found")
	// ErrValidation 表示请求参数不符合要求。
	ErrValidation = errors.New("validation error")
)
