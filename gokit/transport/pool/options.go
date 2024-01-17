package pool

import (
	"errors"
	"net"
	"time"

	jkutils "github.com/jkprj/jkfr/gokit/utils"
)

type PoolClient interface {
	Close() error
}

type ClientFatory func(o *Options) (PoolClient, net.Conn, error)

var (
	ErrClosed   = errors.New("pool closed")
	ErrIniting  = errors.New("Initing")
	ErrInvalid  = errors.New("invalid config")
	ErrRejected = errors.New("connection is nil. rejecting")
	ErrTargets  = errors.New("targets server is empty")
	ErrNoValid  = errors.New("no valid client")
	ErrMax      = errors.New("client max")
	ErrNotFound = errors.New("not found")
)

// Options pool options
type Options struct {
	ServerAddr string
	Codec      string

	// init connection
	InitCap int
	// max connections
	MaxCap       int
	DialTimeout  time.Duration
	IdleTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	Factory ClientFatory `json:"-"`
}

// NewOptions returns a new newOptions instance with sane defaults.
func NewOptions() *Options {
	o := &Options{}
	o.Codec = jkutils.CODEC_GOB
	o.InitCap = 2
	o.MaxCap = 16
	o.DialTimeout = 10 * time.Second
	o.ReadTimeout = 60 * time.Second
	o.WriteTimeout = 60 * time.Second
	o.IdleTimeout = 2 * time.Hour
	return o
}

// validate checks a Config instance.
func (o *Options) validate() error {
	if o.InitCap <= 0 ||
		o.MaxCap <= 0 ||
		o.InitCap > o.MaxCap ||
		o.DialTimeout == 0 ||
		o.ReadTimeout == 0 ||
		o.WriteTimeout == 0 ||
		o.Factory == nil {
		return ErrInvalid
	}
	return nil
}
