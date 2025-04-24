package server

import (
	"context"
	"encoding/base64"
	"strings"

	pb "filesystem/proto/filesystem"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Acá se marca la ruta dentro de la vm donde se ejecuta de donde se guardará cada archivo.
const rootDirectory = "storage"

type Server struct {
	pb.UnimplementedFileSystemServiceServer
}

// // // Registrar el nodo con el servidor central
// func (s *Server) RegisterWithCentral() {
// 	centralAddress := os.Getenv("CENTRAL_SERVER_ADDRESS")
// 	if centralAddress == "" {
// 		log.Fatal("CENTRAL_SERVER_ADDRESS no está definido en el archivo .env")
// 	}
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

// // // Reportar estado del nodo (saber si esta ocupado, activo, etc)
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

// // // Enviar estado cada 10 segundos
// func (s *Server) SendHeartbeat() {
// 	for {
// 		time.Sleep(10 * time.Second)
// 		s.ReportStatus("activo")
// 	}
// }

// Subir archivo en Base64
func (s *Server) UploadFile(ctx context.Context, req *pb.UploadRequest) (*pb.Response, error) {
	filename := req.Filename
	base64Data := req.ContentBase64

	// Cargar variables de entorno desde el archivo .env
	err := godotenv.Load()
	if err != nil {
		log.Println("No se pudo cargar el archivo .env")
	}

	nodeIDStr := os.Getenv("NODE_ID")
	// En caso de estar vacío se pondrá un valor por defecto
	if nodeIDStr == "" {
		nodeIDStr = "1"
	}

	if filename == "" {
		return &pb.Response{Message: "El nombre del archivo no puede estar vacío"}, nil
	}

	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return &pb.Response{Message: "Error decodificando Base64"}, err
	}

	// Crear directorios si no existen
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

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return &pb.Response{Message: "Error escribiendo archivo"}, err
	}

	// Obtener tipo de archivo (MIME type)
	mimeType := http.DetectContentType(data)

	// Alternativamente, usar extensión si deseas más precisión:
	ext := filepath.Ext(filename)
	if ext != "" {
		mimeFromExt := mime.TypeByExtension(ext)
		if mimeFromExt != "" {
			mimeType = mimeFromExt
		}
	}

	return &pb.Response{
		Message:  "Archivo subido correctamente",
		FilePath: filePath,
		FileName: filename,
		FileSize: int64(len(data)),
		FileType: mimeType,
		NodeId:   nodeIDStr,
	}, nil
}

// Mover un archivo
func (s *Server) MoveFile(ctx context.Context, req *pb.MoveRequest) (*pb.Response, error) {
	log.Printf("Datos recibidos:\nSourcePath: %s\nDestinationPath: %s", req.SourcePath, req.DestinationPath)
	sourcePath := filepath.Join(rootDirectory, req.SourcePath)
	destPath := filepath.Join(rootDirectory, req.DestinationPath, filepath.Base(sourcePath))

	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return nil, status.Errorf(codes.NotFound, "El archivo fuente no existe")
	}

	if err := os.Rename(sourcePath, destPath); err != nil {
		return nil, status.Errorf(codes.Internal, "Error al mover archivo: %v", err)
	}
	return &pb.Response{Message: "Archivo movido con éxito"}, nil
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

// Crea un subdirectorio
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

	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return nil, status.Errorf(codes.NotFound, "El archivo a renombrar no existe")
	}

	if err := os.Rename(oldPath, newPath); err != nil {
		return nil, status.Errorf(codes.Internal, "Error al renombrar archivo: %v", err)
	}
	return &pb.Response{Message: "Archivo renombrado con éxito"}, nil
}

// Elimina un archivo o directorio
func (s *Server) DeleteFile(ctx context.Context, req *pb.DeleteRequest) (*pb.Response, error) {
	targetPath := filepath.Join(rootDirectory, req.Path)

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

func (s *Server) ListDirectories(ctx context.Context, req *pb.DirectoryRequest) (*pb.ListResponse, error) {
	fullPath := filepath.Join(rootDirectory, req.Path)

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "No se pudo leer el directorio: %v", err)
	}

	var directories []string
	for _, entry := range entries {
		if entry.IsDir() { // Solo directorios
			directories = append(directories, entry.Name())
		}
	}

	return &pb.ListResponse{Files: directories}, nil
}

func (s *Server) ListAll(ctx context.Context, req *pb.DirectoryRequest) (*pb.ListAllResponse, error) {
	fullPath := filepath.Join(rootDirectory, req.Path)

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "No se pudo leer el directorio: %v", err)
	}

	var files []string
	var directories []string
	for _, entry := range entries {
		if entry.IsDir() {
			directories = append(directories, entry.Name())
		} else {
			files = append(files, entry.Name())
		}
	}

	return &pb.ListAllResponse{
		Files:       files,
		Directories: directories,
	}, nil
}

// Estructura para la respuesta de la descarga
type FileRequest struct {
	Success  bool   `json:"success"`
	FileID   int    `json:"fileID"`
	FileName string `json:"fileName"`
	NodeID   int    `json:"nodeID"`
	NodeIP   string `json:"nodeIP"`
	FilePath string `json:"filePath"`
}

// Descargar un archivo
type fileInfo struct {
	FileName string `json:"fileName"`
	FilePath string `json:"filePath"`
}

func (s *Server) DownloadFile(ctx context.Context, req *pb.DownloadRequest) (*pb.DownloadResponse, error) {
	// Asegúrate de que req.Path sea solo una ruta válida
	if req.Path == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Ruta proporcionada es vacía")
	}

	// Asegurarse de que no esté recibiendo un JSON o algo inesperado
	if strings.Contains(req.Path, "{") || strings.Contains(req.Path, "}") {
		return nil, status.Errorf(codes.InvalidArgument, "La ruta proporcionada contiene un formato JSON inválido: %s", req.Path)
	}

	// Ahora se crea la ruta completa
	fullPath := filepath.Join(rootDirectory, req.Path)

	// Log para saber si se llegó al archivo
	log.Printf("Intentando acceder al archivo: %s", fullPath)

	// Verificar si el archivo existe
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("El archivo no existe: %s", fullPath)
			return nil, status.Errorf(codes.NotFound, "El archivo no existe")
		}
		log.Printf("Error al obtener información del archivo: %v", err)
		return nil, status.Errorf(codes.Internal, "Error al obtener información del archivo: %v", err)
	}

	// Si es un directorio, no un archivo
	if info.IsDir() {
		log.Printf("La ruta proporcionada es un directorio, no un archivo: %s", fullPath)
		return nil, status.Errorf(codes.InvalidArgument, "La ruta proporcionada es un directorio, no un archivo")
	}

	// Log para saber si se encontró el archivo
	log.Printf("Archivo encontrado: %s", fullPath)

	// Leer el archivo
	data, err := os.ReadFile(fullPath)
	if err != nil {
		log.Printf("Error al leer el archivo %s: %v", fullPath, err)
		return nil, status.Errorf(codes.Internal, "Error al leer el archivo: %v", err)
	}
	log.Printf("Tamaño de datos leídos: %d bytes", len(data))

	// Log para saber si el archivo se leyó correctamente
	log.Printf("Archivo leído exitosamente: %s", fullPath)

	// Codificar en Base64
	base64Content := base64.StdEncoding.EncodeToString(data)

	// Log para saber si la conversión a Base64 fue exitosa
	log.Printf("Archivo convertido a Base64 exitosamente: %s", fullPath)

	// Obtener tipo MIME
	mimeType := http.DetectContentType(data)
	ext := filepath.Ext(req.Path)
	if ext != "" {
		mimeFromExt := mime.TypeByExtension(ext)
		if mimeFromExt != "" {
			mimeType = mimeFromExt
		}
	}
	log.Printf("Respuesta enviada al cliente:\nFilename: %s\nFilesize: %d\nFileType: %s\nBase64 (primeros 100): %.100s",
		filepath.Base(fullPath),
		info.Size(),
		mimeType,
		base64Content,
	)

	return &pb.DownloadResponse{
		Filename:      filepath.Base(fullPath),
		ContentBase64: base64Content,
		Filesize:      info.Size(),
		FileType:      mimeType,
	}, nil
}

// Instancia de server
func NewServer() *Server {
	return &Server{}
}
