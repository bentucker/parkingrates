package server

import (
    "fmt"
    "io/ioutil"
    "log"
    "net"
    "strconv"
    "strings"

    pb "github.com/bentucker/parkingrates/genproto"
    "github.com/buger/jsonparser"
    "github.com/golang-collections/go-datastructures/augmentedtree"
    "golang.org/x/net/context"
    "google.golang.org/grpc"
    "github.com/golang/protobuf/ptypes"
    "github.com/gogo/googleapis/google/rpc"
    "google.golang.org/genproto/googleapis/rpc/code"
    "io"
    "os"
)

var weekdays = map[string]string{
    "Monday":    "mon",
    "Tuesday":   "tues",
    "Wednesday": "wed",
    "Thursday":  "thurs",
    "Friday":    "fri",
    "Saturday":  "sat",
    "Sunday":    "sun",
}

type ratesServer struct {
    rates map[string]augmentedtree.Tree
}

type parkingRate struct {
    start, end, price int64
    id                uint64
}

func (ti *parkingRate) LowAtDimension(dimension uint64) int64 {
    return ti.start
}

func (ti *parkingRate) HighAtDimension(dimension uint64) int64 {
    return ti.end
}

//func (ti *parkingRate) ValueAtDimension(dimension uint64) int64 {
//    return ti.price
//}

func (ti *parkingRate) OverlapsAtDimension(interval augmentedtree.Interval,
    dimension uint64) bool {
    return interval.LowAtDimension(dimension) >= ti.LowAtDimension(dimension) &&
        interval.HighAtDimension(dimension) < ti.HighAtDimension(dimension)
}

func (ti *parkingRate) ID() uint64 {
    return ti.id
}

func (ti *parkingRate) PriceAtDimension(dimension uint64) int64 {
    return ti.price
}

func RunServer(cfg string, port int) {
    lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    log.Printf("listening on %s", lis.Addr().String())
    var opts []grpc.ServerOption
    grpcServer := grpc.NewServer(opts...)
    reader, _ := os.Open(cfg)
    defer reader.Close()
    pb.RegisterRatesServer(grpcServer, newServer(reader))
    grpcServer.Serve(lis)
}

func newServer(cfg io.Reader) *ratesServer {
    data, err := ioutil.ReadAll(cfg)
    if err != nil {
        log.Fatalf("Could not read %s", cfg)
    }

    rates, err := parseRates(data)
    if err != nil {
        log.Fatalf("Failed to parse parking rates: %v", err)
    }
    s := &ratesServer{rates: rates}
    return s
}

func addRate(rates map[string]augmentedtree.Tree, value []byte) error {
    daysStr, err := jsonparser.GetString(value, "days")

    if err != nil {
        return err
    }

    timesStr, err := jsonparser.GetString(value, "times")

    if err != nil {
        return err
    }

    price, err := jsonparser.GetInt(value, "price")

    if err != nil {
        return err
    }

    days := strings.Split(daysStr, ",")
    times := strings.Split(timesStr, "-")
    start, err := strconv.ParseInt(times[0], 10, 32)
    if err != nil {
        return err
    }
    end, err := strconv.ParseInt(times[1], 10, 32)
    if err != nil {
        return err
    }

    for _, day := range days {
        if rates[day] == nil {
            rates[day] = augmentedtree.New(1)
        }

        interval := &parkingRate{
            start: start,
            end:   end,
            price: price,
            id:    rates[day].Len(),
        }
        rates[day].Add(interval)
    }

    return nil
}

func parseRates(data []byte) (map[string]augmentedtree.Tree, error) {
    rates := make(map[string]augmentedtree.Tree)

    var parseErr error
    jsonparser.ArrayEach(data,
        func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
            err = addRate(rates, value)
            if err != nil {
                parseErr = err
            }
        }, "rates")

    return rates, parseErr
}

// GetRates returns the parking rates for the given time period
func (s *ratesServer) GetRates(ctx context.Context, req *pb.RatesRequest) (*pb.RatesReply, error) {
    startDate, err := ptypes.Timestamp(req.Start)
    if err != nil {
        return nil, err
    }
    endDate, err := ptypes.Timestamp(req.End)
    if err != nil {
        return nil, err
    }
    startDay := weekdays[startDate.Weekday().String()]
    endDay := weekdays[endDate.Weekday().String()]
    if startDay != endDay {
        // overnight parking is not allowed
        return &pb.RatesReply{
            Status: &rpc.Status{Code: code.Code_value["OUT_OF_RANGE"]},
        }, nil
    }
    startTime := int64(startDate.Hour()*100 + startDate.Minute())
    endTime := int64(endDate.Hour()*100 + endDate.Minute())

    interval := &parkingRate{
        start: startTime,
        end:   endTime,
        price: 0,
    }

    rates := s.rates[startDay].Query(interval)
    var rate *parkingRate = nil
    for _, r := range rates {
        if r.(*parkingRate).start <= startTime && r.(*parkingRate).end > endTime {
            rate = r.(*parkingRate)
            break
        }
    }

    if rate == nil {
        return &pb.RatesReply{
            Status: &rpc.Status{Code: code.Code_value["OUT_OF_RANGE"]},
        }, nil
    }

    return &pb.RatesReply{
        Rate:   float64(rate.price),
        Status: &rpc.Status{Code: code.Code_value["OK"]},
    }, nil
}
