package server

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"github.com/xitongsys/ethernet-go/header"
	"github.com/xitongsys/pangolin/cache"
	"github.com/xitongsys/pangolin/server"
	rediscache "github.com/zedonboy/grize-vpn-server/cache"
	"github.com/zedonboy/grize-vpn-server/signer"
)

type UdpServer struct {
	pangolinServer *server.UdpServer
	RedisCache     *rediscache.RedisCache
}

type Subscription struct {
	SubScriptionId uint32
	// in miliseconds
	Expire int64
}

const secret string = "vfdjvndsdcduicewuifhewifdvnjfvufvuf"
const UDPCHANBUFFERSIZE int = 1024

func NewInstance(serverAddr string, lm *server.LoginManager) (*UdpServer, error) {
	addr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	pserver := &server.UdpServer{
		Addr:          serverAddr,
		UdpConn:       conn,
		LoginManager:  lm,
		TunToConnChan: make(chan string, UDPCHANBUFFERSIZE),
		ConnToTunChan: make(chan string, UDPCHANBUFFERSIZE),
		RouteMap:      cache.NewCache(time.Minute * 10), //cache.NewCache(time.Minute * 10),
	}

	return &UdpServer{
		pangolinServer: pserver,
		RedisCache:     rediscache.NewRedisCache(time.Second * 3600 * 12),
	}, nil
}

func (us *UdpServer) StartWithAuth() {
	us.pangolinServer.LoginManager.TunServer.StartClient("udp", us.pangolinServer.ConnToTunChan, us.pangolinServer.TunToConnChan)
	log.Default().Println("Udp Server Started")
	signer := signer.NewSigner(secret)
	go func() {
		defer func() {
			recover()
		}()

		data := make([]byte, us.pangolinServer.LoginManager.TunServer.TunConn.GetMtu()*2)
		for {
			if n, addr, err := us.pangolinServer.UdpConn.ReadFromUDP(data); err == nil && n > 0 {
				// check data header
				jwtHeader := data[:2]
				count := binary.BigEndian.Uint16(jwtHeader)
				jwtString := data[3:count]
				jwt := string(jwtString)
				jwtFieldList := strings.Split(jwt, ".")

				if len(jwtFieldList) != 2 {
					continue
				}

				if _, ok := us.RedisCache.Get(jwtFieldList[0]); ok {
					us.pangolinServer.ConnToTunChan <- string(data[count+2 : n])
					continue
				}

				if signer.Verify([]byte(jwtFieldList[0]), []byte(jwtFieldList[1])) {
					payload := jwtFieldList[0]
					jsonDataStrng, err := base64.StdEncoding.DecodeString(payload)
					if err != nil {
						continue
					}
					jdec := json.NewDecoder(strings.NewReader(string(jsonDataStrng)))

					var sub Subscription
					for {
						err := jdec.Decode(&sub)
						if err == io.EOF {
							break
						} else if err != nil {
							sub.Expire = -1
							break
						}
					}

					if sub.Expire == -1 {
						continue
					}

					currTime := time.Now()
					elapsed := (sub.Expire / 1000) - currTime.Unix()

					// Subscrition is Valid
					if elapsed > 0 {

						ipHeader := data[count+2 : n]
						proto, src, dst, err := header.GetBase(ipHeader)
						if err != nil {
							continue
						}
						key := proto + ":" + src + ":" + dst
						us.RedisCache.Put(payload, true, sub.Expire)
						us.pangolinServer.RouteMap.Put(key, addr.String())
						us.pangolinServer.ConnToTunChan <- string(ipHeader)
					}

				}

			}
		}
	}()

	go func() {
		defer func() {
			recover()
		}()

		for {
			data, ok := <-us.pangolinServer.TunToConnChan
			if ok {
				if protocol, src, dst, err := header.GetBase([]byte(data)); err == nil {
					key := protocol + ":" + dst + ":" + src
					clientAddrI := us.pangolinServer.RouteMap.Get(key)
					if clientAddrI != nil {
						clientAddr := clientAddrI.(string)
						if add, err := net.ResolveUDPAddr("udp", clientAddr); err == nil {
							us.pangolinServer.UdpConn.WriteToUDP([]byte(data), add)
						}
					}
				}
			}
		}
	}()

}
