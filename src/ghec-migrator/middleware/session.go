package middleware

import (
	"os"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

func SessionStore() *sessions.CookieStore {
	sessionStore := sessions.NewCookieStore(sessionKeys()...)
	sessionStore.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 86400, // 1 day
	}
	return sessionStore
}

func sessionKeys() [][]byte {
	sessionAuthenticationKey := []byte(os.Getenv("SESSION_AUTHENTICATION_KEY"))
	if len(sessionAuthenticationKey) == 0 {
		sessionAuthenticationKey = securecookie.GenerateRandomKey(32)
	}
	sessionEncryptionKey := []byte(os.Getenv("SESSION_ENCRYPTION_KEY"))
	if len(sessionEncryptionKey) == 0 {
		sessionEncryptionKey = securecookie.GenerateRandomKey(32)
	}
	return [][]byte{sessionAuthenticationKey, sessionEncryptionKey}
}
