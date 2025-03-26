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
	// Usar 4 hilos de cpu
	runtime.GOMAXPROCS(4)
	log.Println("Configurado para usar hasta 4 hilos de CPU")

	err := godotenv.Load()
	if err != nil {
		log.Println("Advertencia: No se pudo cargar el archivo .env, usando valores por defecto.")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}

	// pool de workers
	workerCount := 4
	var wg sync.WaitGroup
	workChan := make(chan func(), 100)

	// Iniciar workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			log.Printf("Worker %d iniciado\n", workerID)
			for work := range workChan {
				work()
			}
			log.Printf("Worker %d terminado\n", workerID)
		}(i + 1)
	}

	// Configurar interceptor para el servidor gRPC
	//El handler es el que se encarga de manejar la petición o buscar la función qde lo que se necesita
	unaryInterceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var result interface{}
		var err error
		done := make(chan struct{})

		// Enviar trabajo al pool
		workChan <- func() {
			result, err = handler(ctx, req)
			close(done)
		}

		<-done
		return result, err
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(unaryInterceptor),
	)

	fileSystemServer := server.NewServer()
	pb.RegisterFileSystemServiceServer(grpcServer, fileSystemServer)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Error al escuchar en el puerto %s: %v", port, err)
	}

	log.Printf("Servidor gRPC corriendo en el puerto %s...\n", port)

	// Iniciar el servidor
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Error al iniciar el servidor gRPC: %v", err)
		}
	}()

	waitForShutdown(grpcServer, workChan, &wg)
}

// Para apagar el sv
func waitForShutdown(grpcServer *grpc.Server, workChan chan func(), wg *sync.WaitGroup) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Recibida señal de apagado, cerrando servidor...")

	close(workChan)
	wg.Wait()

	// Detener el servidor gRPC
	grpcServer.GracefulStop()
	log.Println("Servidor detenido correctamente")
}
