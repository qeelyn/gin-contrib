package cache

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"time"
	"net/url"
	"crypto/sha1"
	"io"
	"bytes"
	"sync"
	"github.com/qeelyn/go-common/cache"
)

const (
	CACHE_MIDDLEWARE_KEY = "gincontrib.cache"
)

var(
	PageCachePrefix = "gincontrib.page.cache"
)

type responseCache struct {
	Status int
	Header http.Header
	Data   []byte
}

type cachedWriter struct {
	gin.ResponseWriter
	status  int
	written bool
	store   cache.Cache
	expire  time.Duration
	key     string
}

var _ gin.ResponseWriter = &cachedWriter{}

func urlEscape(prefix string, u string) string {
	key := url.QueryEscape(u)
	if len(key) > 200 {
		h := sha1.New()
		io.WriteString(h, u)
		key = string(h.Sum(nil))
	}
	var buffer bytes.Buffer
	buffer.WriteString(prefix)
	buffer.WriteString(":")
	buffer.WriteString(key)
	return buffer.String()
}

func newCachedWriter(store cache.Cache, expire time.Duration, writer gin.ResponseWriter, key string) *cachedWriter {
	return &cachedWriter{writer, 0, false, store, expire, key}
}

func (w *cachedWriter) WriteHeader(code int) {
	w.status = code
	w.written = true
	w.ResponseWriter.WriteHeader(code)
}

func (w *cachedWriter) Status() int {
	return w.ResponseWriter.Status()
}

func (w *cachedWriter) Written() bool {
	return w.ResponseWriter.Written()
}

func (w *cachedWriter) Write(data []byte) (int, error) {
	ret, err := w.ResponseWriter.Write(data)
	if err == nil {
		store := w.store
		var rc responseCache
		if err := store.Get(w.key, &rc); err == nil {
			data = append(rc.Data, data...)
		}

		//rc response
		val := responseCache{
			w.status,
			w.Header(),
			data,
		}
		err = store.Set(w.key, val, w.expire)
		if err != nil {
			// need logger
		}
	}
	return ret, err
}

func (w *cachedWriter) WriteString(data string) (n int, err error) {
	ret, err := w.ResponseWriter.WriteString(data)
	if err == nil {
		//cache response
		store := w.store
		val := responseCache{
			w.ResponseWriter.Status(),
			w.Header(),
			[]byte(data),
		}
		store.Set(w.key, val, w.expire)
	}
	return ret, err
}

// Cache Middleware
func CacheHandle(store cache.Cache) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(CACHE_MIDDLEWARE_KEY, store)
		c.Next()
	}
}

func SiteCacheHandle(store cache.Cache, expire time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		var rc responseCache
		rRrl := c.Request.URL
		key := urlEscape(PageCachePrefix, rRrl.RequestURI())
		if err := store.Get(key, &rc); err != nil {
			c.Next()
		} else {
			c.Writer.WriteHeader(rc.Status)
			for k, vals := range rc.Header {
				for _, v := range vals {
					c.Writer.Header().Add(k, v)
				}
			}
			c.Writer.Write(rc.Data)
		}
	}
}

// CachePage Decorator
func CachePageHandle(store cache.Cache, expire time.Duration, handle gin.HandlerFunc) gin.HandlerFunc {

	return func(c *gin.Context) {
		var rc responseCache
		rRrl := c.Request.URL
		key := urlEscape(PageCachePrefix, rRrl.RequestURI())
		if store.IsExist(key) {
			if err := store.Get(key, &rc); err == nil {
				c.Writer.WriteHeader(rc.Status)
				for k, vals := range rc.Header {
					for _, v := range vals {
						c.Writer.Header().Add(k, v)
					}
				}
				c.Writer.Write(rc.Data)
				return
			}
		}
		// replace writer
		writer := newCachedWriter(store, expire, c.Writer, key)
		c.Writer = writer
		handle(c)
	}
}

// CachePageAtomic Decorator
func CachePageAtomicHandle(store cache.Cache, expire time.Duration, handle gin.HandlerFunc) gin.HandlerFunc {
	var m sync.Mutex
	p := CachePageHandle(store, expire, handle)
	return func(c *gin.Context) {
		m.Lock()
		defer m.Unlock()
		p(c)
	}
}

func CachePageWithoutHeaderHandle(store cache.Cache, expire time.Duration, handle gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		var rc responseCache
		rRrl := c.Request.URL
		key := urlEscape(PageCachePrefix, rRrl.RequestURI())
		if store.IsExist(key) {
			if err := store.Get(key, &rc); err == nil {
				c.Writer.WriteHeader(rc.Status)
				c.Writer.Write(rc.Data)
				return
			}
		}
		writer := newCachedWriter(store, expire, c.Writer, key)
		c.Writer = writer
		handle(c)
	}
}

