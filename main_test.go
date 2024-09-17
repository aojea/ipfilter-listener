package main

import (
	"net"
	"net/netip"
	"testing"
)

func TestIPFilterListener(t *testing.T) {
	tests := []struct {
		name     string
		ipBlocks []netip.Prefix
		want     bool
	}{
		{
			name: "deny all",
			want: false,
		},
		{
			name:     "allow from localhost",
			ipBlocks: []netip.Prefix{netip.MustParsePrefix("127.0.0.0/8")},
			want:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			closeCh := make(chan struct{})
			l, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatalf("unexpected error %v creating listener", err)
			}
			filteredListener := IPFilterListener(l, tt.ipBlocks)
			defer filteredListener.Close()

			go func() {
				for {
					c, err := filteredListener.Accept()
					if err != nil {
						if tt.want {
							t.Errorf("expected connection to succeed, got %v", err)
						}
						close(closeCh)
						return
					}
					if !tt.want {
						t.Errorf("expected connection from %s to fail", c.RemoteAddr().String())
					}
					t.Logf("connection from %s succeeded", c.RemoteAddr().String())
					close(closeCh)
					return
				}
			}()
			conn, err := net.Dial(filteredListener.Addr().Network(), filteredListener.Addr().String())
			defer conn.Close()
			<-closeCh

		})
	}
}
