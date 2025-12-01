package request

type UploadKnowledgeMetadataRequest struct {
	FileName   string `json:"file_name"`
	FileType   string `json:"file_type"`
	FileSize   int64  `json:"file_size"`
	ObjectName string `json:"object_name"`
}
