package server

import (
	"context"
	"os"

	"path/filepath"

	pb "filesystem/proto/filesystem"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Archivo raíz (el default del user)
const rootDirectory = "storage"

type Server struct {
	pb.UnimplementedFileSystemServiceServer
}

// // Registrar el nodo con el servidor central
// func (s *Server) RegisterWithCentral() {
// centralAddress := os.Getenv("CENTRAL_SERVER_ADDRESS")
// if centralAddress == "" {
// 	log.Fatal("CENTRAL_SERVER_ADDRESS no está definido en el archivo .env")
// }
// 	conn, err := grpc.Dial("centralAddress", grpc.WithInsecure())
// 	if err != nil {
// 		log.Fatalf("No se pudo conectar al servidor central: %v", err)
// 	}
// 	defer conn.Close()

// 	client := pb.NewNodeServiceClient(conn)
// 	_, err = client.RegisterNode(context.Background(), &pb.NodeInfo{
// 		Address: "nodo-direccion",
// 		Status:  "activo",
// 	})
// 	if err != nil {
// 		log.Fatalf("Error registrando el nodo: %v", err)
// 	}
// }

// // Reportar estado del nodo (saber si esta ocupado, activo, etc)
// func (s *Server) ReportStatus(status string) {
// 	conn, err := grpc.Dial("central-server-address", grpc.WithInsecure())
// 	if err != nil {
// 		log.Fatalf("No se pudo conectar al servidor central: %v", err)
// 	}
// 	defer conn.Close()

// 	client := pb.NewNodeServiceClient(conn)
// 	_, err = client.ReportStatus(context.Background(), &pb.NodeStatus{
// 		Address: "nodo-direccion",
// 		Status:  status,
// 	})
// 	if err != nil {
// 		log.Fatalf("Error reportando estado: %v", err)
// 	}
// }

// // Enviar estado cada 10 segundos
// func (s *Server) SendHeartbeat() {
// 	for {
// 		time.Sleep(10 * time.Second)
// 		s.ReportStatus("activo")
// 	}
// }

// Subir archivo
func (s *Server) UploadFile(ctx context.Context, req *pb.UploadRequest) (*pb.Response, error) {
	filename := req.Filename
	data := req.Content

	// Verificar que no esté vacío
	if filename == "" {
		return &pb.Response{Message: "El nombre del archivo no puede estar vacío"}, nil
	}

	if err := os.MkdirAll(rootDirectory, os.ModePerm); err != nil {
		return &pb.Response{Message: "Error creando directorio raíz"}, err
	}

	filePath := filepath.Join(rootDirectory, filename)
	if req.Directory != "" {
		fullDir := filepath.Join(rootDirectory, req.Directory)
		if err := os.MkdirAll(fullDir, os.ModePerm); err != nil {
			return &pb.Response{Message: "Error creando directorio especificado"}, err
		}
		filePath = filepath.Join(fullDir, filename)
	}

	err := os.WriteFile(filePath, data, 0644)
	if err != nil {
		return &pb.Response{Message: "Error escribiendo archivo"}, err
	}

	return &pb.Response{Message: "Archivo subido correctamente a " + filePath}, nil
}

// Mover un archivo
func (s *Server) MoveFile(ctx context.Context, req *pb.MoveRequest) (*pb.Response, error) {
	sourcePath := req.SourcePath
	destinationPath := req.DestinationPath

	if _, err := os.Stat(req.SourcePath); os.IsNotExist(err) {
		return nil, status.Errorf(codes.NotFound, "El archivo '%s' no existe en el servidor", req.SourcePath)
	}

	// Mover el archivo o directorio
	err := os.Rename(sourcePath, destinationPath)
	if err != nil {
		return &pb.Response{Message: "Error moviendo archivo/directorio"}, err
	}
	return &pb.Response{Message: "Archivo/Directorio movido correctamente"}, nil
}

// Crear un nuevo directorio
func (s *Server) CreateDirectory(ctx context.Context, req *pb.DirectoryRequest) (*pb.Response, error) {
	if req.Path == "" {
		return &pb.Response{Message: "La ruta del directorio no puede estar vacía"}, nil
	}

	fullPath := filepath.Join(rootDirectory, req.Path)
	err := os.MkdirAll(fullPath, os.ModePerm)
	if err != nil {
		return &pb.Response{Message: "Error creando directorio"}, err
	}
	return &pb.Response{Message: "Directorio creado correctamente"}, nil
}

// Crea un subdirectorio dentro de otro directorio
func (s *Server) CreateSubdirectory(ctx context.Context, req *pb.SubdirectoryRequest) (*pb.Response, error) {
	if req.ParentDirectory == "" || req.SubdirectoryName == "" {
		return &pb.Response{Message: "El nombre del directorio padre y del subdirectorio no pueden estar vacíos"}, nil
	}

	fullPath := filepath.Join(rootDirectory, req.ParentDirectory, req.SubdirectoryName)
	err := os.MkdirAll(fullPath, os.ModePerm)
	if err != nil {
		return &pb.Response{Message: "Error creando subdirectorio"}, err
	}
	return &pb.Response{Message: "Subdirectorio creado correctamente"}, nil
}

// Renombra un archivo o directorio
func (s *Server) RenameFile(ctx context.Context, req *pb.RenameRequest) (*pb.Response, error) {
	oldPath := filepath.Join(rootDirectory, req.OldName)
	newPath := filepath.Join(rootDirectory, req.NewName)

	// Verificar si el archivo/directorio existe antes de renombrar
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return &pb.Response{Message: "El archivo/directorio no existe"}, nil
	}

	err := os.Rename(oldPath, newPath)
	if err != nil {
		return &pb.Response{Message: "Error renombrando archivo/directorio"}, err
	}
	return &pb.Response{Message: "Archivo/Directorio renombrado correctamente"}, nil
}

// Elimina un archivo o directorio
func (s *Server) DeleteFile(ctx context.Context, req *pb.DeleteRequest) (*pb.Response, error) {
	targetPath := filepath.Join(rootDirectory, req.Path)

	// Verificar si el archivo/directorio existe
	info, err := os.Stat(targetPath)
	if os.IsNotExist(err) {
		return &pb.Response{Message: "El archivo/directorio no existe"}, nil
	}

	if info.IsDir() {
		// Eliminar el directorio y su contenido
		err = os.RemoveAll(targetPath)
		if err != nil {
			return &pb.Response{Message: "Error eliminando directorio"}, err
		}
		return &pb.Response{Message: "Directorio eliminado correctamente"}, nil
	}

	// Eliminar archivo
	err = os.Remove(targetPath)
	if err != nil {
		return &pb.Response{Message: "Error eliminando archivo"}, err
	}
	return &pb.Response{Message: "Archivo eliminado correctamente"}, nil
}

// Lista los archivos de un directorio
func (s *Server) ListFiles(ctx context.Context, req *pb.DirectoryRequest) (*pb.ListResponse, error) {
	fullPath := filepath.Join(rootDirectory, req.Path)

	// Verificar si el directorio existe
	files, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	var filenames []string
	for _, file := range files {
		filenames = append(filenames, file.Name())
	}

	return &pb.ListResponse{Files: filenames}, nil
}

// NewServer crea una nueva instancia del servidor gRPC
func NewServer() *Server {
	return &Server{}
}
