package main

import (
	"context"
	"net"
)

type contextKey = string

const usernameKey contextKey = "username"

type UserConn struct {
	Conn     net.Conn
	Username string
	Ctx      context.Context
}

var activeConns = make(map[string]*UserConn)

func WithUsername(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, usernameKey, username)
}

func GetUsernameFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(usernameKey).(string)
	return username, ok
}