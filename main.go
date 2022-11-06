package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"

	"github.com/usrmaia/RPC-Server/pb"
	"google.golang.org/grpc"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

type Server struct {
	pb.UnimplementedSendMessageServer
}

type Part struct {
	Id    int64   `json:"id"`
	Name  string  `json:"name"`
	Brand string  `json:"brand"`
	Value float32 `json:"value"`
}

func (Server *Server) Home(ctx context.Context, req *pb.RequestMessage) (*pb.ResponseStatus, error) {
	res := &pb.ResponseStatus{
		Status: "OK",
	}

	return res, nil
}

func (Server *Server) OpenDB(ctx context.Context, req *pb.RequestDataSourceName) (*pb.ResponseStatus, error) {
	var err error
	DB, err = sql.Open("mysql", req.GetDataSourceName())

	if err != nil {
		log.Fatal("Error Open Database: ", err.Error())
		return nil, err
	}

	err = DB.Ping()

	if err != nil {
		log.Fatal("Error Ping Database: ", err.Error())
		return nil, err
	}

	_, err = DB.Exec(`
	create table if not exists Part (
			id int not null auto_increment,
			name varchar(500) not null unique,
			brand varchar(50) not null,
			value float not null,
			primary key (id)
		)
		`)

	if err != nil {
		log.Fatal("Error Exec CREATE Database: ", err.Error())
		return nil, err
	}

	res := &pb.ResponseStatus{
		Status: "Connected",
	}

	fmt.Println("Conected: mysql", req.GetDataSourceName())

	return res, nil
}

func (Server *Server) ReturnAPart(ctx context.Context, req *pb.RequestPartID) (*pb.ResponsePart, error) {
	var row *sql.Row
	row = DB.QueryRow(`
		select id, name, brand, value 
		from Part
		where id = ?
	`, req.GetId())

	var part Part
	err := row.Scan(&part.Id, &part.Name, &part.Brand, &part.Value)

	if err != nil {
		fmt.Println("Error Scan QueryRow Database:", err)
		return nil, err
	}

	res := &pb.ResponsePart{
		Id:    part.Id,
		Name:  part.Name,
		Brand: part.Brand,
		Value: part.Value,
	}

	return res, nil
}

func (Server *Server) ReturnParts(ctc context.Context, req *pb.RequestMessage) (*pb.ResponseParts, error) {
	var rows *sql.Rows
	var err error
	rows, err = DB.Query(`select id, name, brand, value from Part`)

	if err != nil {
		fmt.Println("Error Query Database: ", err)
		return nil, err
	}

	var Parts []Part
	for rows.Next() {
		var part Part
		err = rows.Scan(&part.Id, &part.Name, &part.Brand, &part.Value)

		if err != nil {
			continue
		}

		Parts = append(Parts, part)
	}

	err = rows.Close()

	if err != nil {
		fmt.Println("Error Close Rows Database:", err)
		return nil, err
	}

	var res *pb.ResponseParts
	res = &pb.ResponseParts{}

	for _, part := range Parts {
		res.Parts = append(res.Parts, &pb.ResponseParts_Part{
			Id:    part.Id,
			Name:  part.Name,
			Brand: part.Brand,
			Value: part.Value,
		})
	}

	return res, nil
}

func (Server *Server) AddPart(ctx context.Context, req *pb.RequestAdd) (*pb.ResponsePart, error) {
	var result sql.Result
	result, err := DB.Exec(`
		insert into Part (name, brand, value) values
		(?, ?, ?)
	`, req.GetName(), req.GetBrand(), req.GetValue())

	if err != nil {
		fmt.Println("Error Exec INSERT Part:", err)
		return nil, err
	}

	var id int64
	id, err = result.LastInsertId()

	if err != nil {
		fmt.Println("Error LastInsertId:", err)
		return nil, err
	}

	res := &pb.ResponsePart{
		Id:    id,
		Name:  req.GetName(),
		Brand: req.GetBrand(),
		Value: req.GetValue(),
	}

	return res, nil
}

func (Server *Server) DelPart(ctx context.Context, req *pb.RequestPartID) (*pb.ResponsePart, error) {
	row := DB.QueryRow(`
		select id, name, brand, value
		from Part
		where id = ? 
	`, req.GetId())

	var err error
	var temp_part Part
	err = row.Scan(&temp_part.Id, &temp_part.Name, &temp_part.Brand, &temp_part.Value)

	if err != nil {
		fmt.Println("Error Scan QueryRow Database:", err)
		return nil, err
	}

	_, err = DB.Exec(`
		delete from Part
		where id = ?
	`, req.GetId())

	if err != nil {
		fmt.Println("Error Exec DELETE Part:", err)
		return nil, err
	}

	res := &pb.ResponsePart{
		Id:    int64(temp_part.Id),
		Name:  temp_part.Name,
		Brand: temp_part.Brand,
		Value: temp_part.Value,
	}

	return res, nil
}

func (Server *Server) UpPart(ctx context.Context, req *pb.RequestUp) (*pb.ResponsePart, error) {
	//exist?
	row := DB.QueryRow(`
		select id
		from Part
		where id = ?
	`, req.GetId())

	var part_id int
	err := row.Scan(&part_id)

	if err != nil {
		fmt.Println("Error Scan QueryRow Database:", err)
		return nil, err
	}

	//update
	_, err = DB.Exec(`
		update Part 
		set name = ?, brand = ?, value = ?
		where id = ?
	`, req.GetName(), req.GetBrand(), req.GetValue(), req.GetId())

	if err != nil {
		fmt.Println("Error Exec UPDATE Part Database:", err)
		return nil, err
	}

	res := &pb.ResponsePart{
		Id:    req.GetId(),
		Name:  req.GetName(),
		Brand: req.GetBrand(),
		Value: req.GetValue(),
	}

	return res, nil
}

func main() {
	var gRPCServer *grpc.Server
	gRPCServer = grpc.NewServer()

	pb.RegisterSendMessageServer(gRPCServer, &Server{})

	address := ":9091"
	var listener net.Listener
	var err error
	listener, err = net.Listen("tcp", address)

	if err != nil {
		fmt.Println("Error Listen Server gRPC:", err)
		log.Fatal(err)
	}

	fmt.Println("Server RPC On in", address)
	err = gRPCServer.Serve(listener)

	if err != nil {
		fmt.Println("Error Serve Server gRPC:", err)
		log.Fatal(err)
	}
}
