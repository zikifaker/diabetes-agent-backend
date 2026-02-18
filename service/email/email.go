package email

import (
	"crypto/tls"
	"net"
	"net/smtp"
)

func Send(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	host, _, _ := net.SplitHostPort(addr)

	// 建立与 SMTP 服务器的 TLS 连接
	conn, err := tls.Dial("tcp", addr, nil)
	if err != nil {
		return err
	}

	// 创建一个新的SMTP客户端
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Quit()

	// 若 SMTP 服务器支持身份验证，则进行验证
	if auth != nil {
		if ok, _ := client.Extension("AUTH"); ok {
			if err := client.Auth(auth); err != nil {
				return err
			}
		}
	}

	// 设置发件方
	if err := client.Mail(from); err != nil {
		return err
	}

	// 设置收件方
	for _, addr := range to {
		if err := client.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	// 写入邮件消息
	if _, err := w.Write(msg); err != nil {
		return err
	}

	return w.Close()
}
