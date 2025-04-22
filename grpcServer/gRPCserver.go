package main

import (
	"context"
	pb "filesystem/proto/filesystem"
	"filesystem/server"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No se pudo cargar el archivo .env, usando valores por defecto.")
	}

	ip := os.Getenv("IP_ADDRESS")
	if ip == "" {
		ip = "127.0.0.1"
		log.Println("IP_ADDRESS no definido, usando 127.0.0.1 por defecto.")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
		log.Println("PORT no definido, usando 50051 por defecto.")
	}

	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		nodeID = "1"
		log.Println("NODE_ID no definido, usando 1 por defecto.")
	}

	log.Printf("Nodo %s iniciando en %s:%s\n", nodeID, ip, port)

	runtime.GOMAXPROCS(runtime.NumCPU())

	var wg sync.WaitGroup
	workChan := make(chan func(), 100)

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			log.Printf("Worker %d iniciado\n", workerID)
			for work := range workChan {
				work()
			}
			log.Printf("Worker %d detenido\n", workerID)
		}(i + 1)
	}

	unaryInterceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var result interface{}
		var err error
		done := make(chan struct{})

		workChan <- func() {
			result, err = handler(ctx, req)
			close(done)
		}

		<-done
		return result, err
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(unaryInterceptor),
		grpc.MaxRecvMsgSize(20*1024*1024),
		grpc.MaxSendMsgSize(20*1024*1024),
	)

	fileSystemServer := server.NewServer()
	pb.RegisterFileSystemServiceServer(grpcServer, fileSystemServer)

	address := ip + ":" + port
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Error al escuchar en %s: %v", address, err)
	}

	log.Printf("Servidor gRPC corriendo en %s\n", address)

	go func() {
		for {
			conn, err := lis.Accept()
			if err != nil {
				log.Printf("Error aceptando conexión: %v", err)
				continue
			}
			log.Printf("Nueva conexión desde: %s", conn.RemoteAddr().String())
			conn.Close()
		}
	}()

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Error al iniciar el servidor gRPC: %v", err)
		}
	}()

	waitForShutdown(grpcServer, workChan, &wg)
}

func waitForShutdown(grpcServer *grpc.Server, workChan chan func(), wg *sync.WaitGroup) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Señal de apagado recibida, cerrando servidor...")

	close(workChan)
	wg.Wait()

	grpcServer.GracefulStop()
	log.Println("Servidor detenido correctamente")
}
