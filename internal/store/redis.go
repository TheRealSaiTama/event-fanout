package store

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type Redis struct { *redis.Client }

func NewRedis(addr, pass string) *Redis {
	c := redis.NewClient(&redis.Options{ Addr: addr, Password: pass, DB: 0 })
	return &Redis{c}
}

func Health(ctx context.Context, r *Redis) error { return r.Ping(ctx).Err() }

func RoomChannel(room string) string { return "room:" + room }
func ParseRoom(channel string) string {
	const p = "room:"
	if len(channel) >= len(p) && channel[:len(p)] == p { return channel[len(p):] }
	return ""
}
