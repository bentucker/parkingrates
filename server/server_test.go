package server

import (
    "reflect"
    "testing"

    pb "github.com/bentucker/parkingrates/genproto"
    "github.com/golang-collections/go-datastructures/augmentedtree"
    "github.com/stretchr/testify/require"
    "golang.org/x/net/context"
    "strings"
    "time"
    "github.com/golang/protobuf/ptypes"
    "github.com/gogo/googleapis/google/rpc"
)

func mkParkingRate(start int64, end int64, price int64) augmentedtree.Interval {
    return &parkingRate{
        start: start,
        end:   end,
        price: price,
    }
}

func TestParseRates(t *testing.T) {
    tests := []struct {
        name    string
        rates   []byte
        want    map[string]augmentedtree.Interval
        wantErr bool
    }{
        {
            name: "Parse single day",
            rates: []byte(`{
                    "rates": [{
                                "days": "mon",
                                "times": "0900-1200",
                                "price": 1000,
                            }]
                }"`),
            want:    map[string]augmentedtree.Interval{"mon": mkParkingRate(900, 1200, 1000)},
            wantErr: false,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := parseRates(tt.rates)
            if (err != nil) != tt.wantErr {
                t.Errorf("parseRates() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            require.Equal(t, len(tt.want), len(got))
            for day := range tt.want {
                entries := got[day].Query(tt.want[day])
                require.Equal(t, len(entries), 1)
            }
        })
    }
}

func Test_ratesServer_GetRates(t *testing.T) {
    type args struct {
        ctx context.Context
        req *pb.RatesRequest
    }
    cfg := `{"rates":
            [{
              "days": "mon,tues,wed,thurs,fri,sat,sun",
              "times": "0900-2000",
              "price": 1500
            }]
            }`
    reader := strings.NewReader(cfg)
    server := newServer(reader)
    start, _ := time.Parse(time.RFC3339, "2015-07-01T09:00:00Z")
    end1, _ := time.Parse(time.RFC3339, "2015-07-01T20:00:00Z")
    end2, _ := time.Parse(time.RFC3339, "2015-07-01T19:00:00Z")

    pbStart, _ := ptypes.TimestampProto(start)
    pbEnd1, _ := ptypes.TimestampProto(end1)
    pbEnd2, _ := ptypes.TimestampProto(end2)

    tests := []struct {
        name    string
        s       *ratesServer
        args    args
        want    *pb.RatesReply
        wantErr bool
    }{
        {
            name: "Invalid range returns error",
            s:    server,
            args: args{
                ctx: context.Background(),
                req: &pb.RatesRequest{
                    Start: pbStart,
                    End:   pbEnd1,
                },
            },
            want: &pb.RatesReply{
                Status: &rpc.Status{Code: 11},
            },
        },
        {
            name: "Correct rate is returned for valid range",
            s:    server,
            args: args{
                ctx: context.Background(),
                req: &pb.RatesRequest{
                    Start: pbStart,
                    End:   pbEnd2,
                },
            },
            want: &pb.RatesReply{
                Status: &rpc.Status{Code: 0},
                Rate:   1500,
            },
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := tt.s.GetRates(tt.args.ctx, tt.args.req)
            if (err != nil) != tt.wantErr {
                t.Errorf("ratesServer.GetRates() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("ratesServer.GetRates() = %v, want %v", got, tt.want)
            }
        })
    }
}
