package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/myhaiting/go-fly-lib/middlewares/keyauth"
	"github.com/myhaiting/go-fly-lib/satoken"
	"github.com/myhaiting/go-fly-lib/satoken/store"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cast"
)

func main() {
	h := server.Default()
	mgr := satoken.NewDefaultManager()
	mgr.MapTokenStorage(store.NewRedisStore(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	}))

	h.Use(keyauth.New(
		keyauth.WithFilter(func(ctx context.Context, c *app.RequestContext) bool {
			if string(c.Request.RequestURI()) == "/login" {
				return true
			}
			return false
		}),
		keyauth.WithManager(mgr),
	),
	)

	// 无需验证
	h.GET("/login", func(ctx context.Context, c *app.RequestContext) {
		token, err := mgr.Login(ctx, "123", satoken.LoginModel{})
		if err != nil {
			c.WriteString("error:" + err.Error())
			return
		}
		sess, err := mgr.GetSession(ctx, token, true)
		if err != nil {
			c.WriteString("error:" + err.Error())
			return
		}
		sess.Set("demo", "11111")
		sess.Save()

		c.WriteString(fmt.Sprintf("token: %s", token))
	})

	h.GET("/logout", func(ctx context.Context, c *app.RequestContext) {
		token, _ := c.Get("token")
		err := mgr.LogoutByToken(ctx, cast.ToString(token))
		if err != nil {
			c.WriteString("error:" + err.Error())
			return
		}
		c.WriteString("ok")
	})

	// 需要token验证
	h.GET("/private", func(ctx context.Context, c *app.RequestContext) {
		token, _ := c.Get("token")
		loginId, _ := c.Get("loginId")
		fmt.Printf("token: %v loginId: %v\n", token, loginId)

		sess, err := mgr.GetSession(ctx, cast.ToString(token), true)
		if err != nil {
			c.WriteString("error:" + err.Error())
			return
		}

		fmt.Printf("session: %v\n", sess.Get("demo"))

		c.WriteString("ok")
	})

	h.Spin()
}
