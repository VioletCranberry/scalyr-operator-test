package clients

type PutFileResponse struct {
	ApiResponse
	Path string `json:"path"`
}

type PutFileRequest struct {
	Path            string `json:"path"`
	Content         string `json:"content"`
	PrettyPrint     bool   `json:"prettyprint"`
	ExpectedVersion int64  `json:"expectedVersion,omitempty"`
}

func (client *ApiClient) PutFile(
	filePath string,
	fileContent string) (*PutFileResponse, error) {
	request := &PutFileRequest{
		Path:        filePath,
		Content:     fileContent,
		PrettyPrint: true,
	}
	response, err := client.PostRequest(
		"/api/putFile",
		request, &PutFileResponse{})
	return response.Result().(*PutFileResponse), err
}

type DeleteFileResponse struct {
	ApiResponse
}

type DeleteFileRequest struct {
	Path            string `json:"path"`
	DeleteFile      bool   `json:"deleteFile"`
	ExpectedVersion int64  `json:"expectedVersion"`
}

func (client *ApiClient) DeleteFile(
	filePath string) (*DeleteFileResponse, error) {
	request_body := &DeleteFileRequest{
		Path:       filePath,
		DeleteFile: true,
	}
	response, err := client.PostRequest(
		"/api/putFile",
		request_body, &DeleteFileResponse{})
	return response.Result().(*DeleteFileResponse), err
}
