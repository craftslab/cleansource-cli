package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"

	"github.com/craftslab/cleansource-sca-cli/internal/logger"
	"github.com/craftslab/cleansource-sca-cli/internal/model"
)

// RemotingClient handles communication with the remote server
type RemotingClient struct {
	client    *resty.Client
	serverURL string
	log       *logrus.Logger
	authToken string
	cookies   []*http.Cookie
}

// NewRemotingClient creates a new remoting client
func NewRemotingClient(serverURL string) *RemotingClient {
	client := resty.New()
	client.SetTimeout(30 * time.Minute) // Long timeout for file uploads
	client.SetRetryCount(3)
	client.SetRetryWaitTime(5 * time.Second)

	return &RemotingClient{
		client:    client,
		serverURL: serverURL,
		log:       logger.GetLogger(),
	}
}

// Login authenticates with username and password
func (rc *RemotingClient) Login(username, password string) error {
	loginData := map[string]string{
		"username": username,
		"password": password,
	}

	resp, err := rc.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(loginData).
		Post(rc.serverURL + "/api/auth/login")

	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	// Store cookies for future requests
	rc.cookies = resp.Cookies()
	rc.log.Info("Login successful")
	return nil
}

// VerifyToken verifies an authentication token
func (rc *RemotingClient) VerifyToken(token string) error {
	resp, err := rc.client.R().
		SetHeader("Authorization", "Bearer "+token).
		Get(rc.serverURL + "/api/auth/verify")

	if err != nil {
		return fmt.Errorf("token verification request failed: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("token verification failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	rc.authToken = token
	rc.log.Info("Token verification successful")
	return nil
}

// UploadData uploads scan data to the server
func (rc *RemotingClient) UploadData(uploadData *model.UploadData) (bool, error) {
	rc.log.Info("Starting data upload...")

	// Create multipart form
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add files
	if err := rc.addFileToForm(writer, "wfpFile", uploadData.WfpFile); err != nil {
		return false, fmt.Errorf("failed to add wfp file: %w", err)
	}

	if uploadData.BuildFile != "" {
		if err := rc.addFileToForm(writer, "buildFile", uploadData.BuildFile); err != nil {
			return false, fmt.Errorf("failed to add build file: %w", err)
		}
	}

	if uploadData.ArchiveFile != "" {
		if err := rc.addFileToForm(writer, "archiveFile", uploadData.ArchiveFile); err != nil {
			return false, fmt.Errorf("failed to add archive file: %w", err)
		}
	}

	// Add metadata
	metadata := rc.createUploadMetadata(uploadData)
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return false, fmt.Errorf("failed to serialize metadata: %w", err)
	}

	if err := writer.WriteField("metadata", string(metadataJSON)); err != nil {
		return false, fmt.Errorf("failed to add metadata: %w", err)
	}

	_ = writer.Close()

	// Create request
	req := rc.client.R().
		SetHeader("Content-Type", writer.FormDataContentType()).
		SetBody(requestBody.Bytes())

	// Add authentication
	if rc.authToken != "" {
		req.SetHeader("Authorization", "Bearer "+rc.authToken)
	} else if len(rc.cookies) > 0 {
		for _, cookie := range rc.cookies {
			req.SetCookie(cookie)
		}
	}

	// Send request
	resp, err := req.Post(rc.serverURL + "/api/scan/upload")
	if err != nil {
		return false, fmt.Errorf("upload request failed: %w", err)
	}

	if resp.StatusCode() != 200 {
		return false, fmt.Errorf("upload failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	// Parse response
	var result model.ScanResult
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		rc.log.Warnf("Failed to parse upload response: %v", err)
		// Assume success if we can't parse the response but got 200
		return true, nil
	}

	rc.log.Infof("Upload completed. Task ID: %s", result.TaskID)
	return result.Success, nil
}

// addFileToForm adds a file to the multipart form
func (rc *RemotingClient) addFileToForm(writer *multipart.Writer, fieldName, filePath string) error {
	if filePath == "" {
		return nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	part, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
	if err != nil {
		return err
	}

	_, err = io.Copy(part, file)
	return err
}

// createUploadMetadata creates metadata for the upload
func (rc *RemotingClient) createUploadMetadata(uploadData *model.UploadData) map[string]interface{} {
	cfg := uploadData.Config

	metadata := map[string]interface{}{
		"taskType":    cfg.TaskType,
		"scanType":    cfg.ScanType,
		"dirSize":     uploadData.DirSize,
		"buildDepend": cfg.BuildDepend,
	}

	if cfg.CustomProject != "" {
		metadata["customProject"] = cfg.CustomProject
	}
	if cfg.CustomProduct != "" {
		metadata["customProduct"] = cfg.CustomProduct
	}
	if cfg.CustomVersion != "" {
		metadata["customVersion"] = cfg.CustomVersion
	}
	if cfg.LicenseName != "" {
		metadata["licenseName"] = cfg.LicenseName
	}
	if cfg.NotificationEmail != "" {
		metadata["notificationEmail"] = cfg.NotificationEmail
	}

	return metadata
}

// VerifyLicense verifies a license name with the server
func (rc *RemotingClient) VerifyLicense(licenseName string) error {
	req := rc.client.R().
		SetQueryParam("licenseName", licenseName)

	// Add authentication
	if rc.authToken != "" {
		req.SetHeader("Authorization", "Bearer "+rc.authToken)
	} else if len(rc.cookies) > 0 {
		for _, cookie := range rc.cookies {
			req.SetCookie(cookie)
		}
	}

	resp, err := req.Get(rc.serverURL + "/api/license/verify")
	if err != nil {
		return fmt.Errorf("license verification request failed: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("license verification failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	rc.log.Info("License verification successful")
	return nil
}

// VerifyEmail verifies an email address with the server
func (rc *RemotingClient) VerifyEmail(email string) error {
	req := rc.client.R().
		SetQueryParam("email", email)

	// Add authentication
	if rc.authToken != "" {
		req.SetHeader("Authorization", "Bearer "+rc.authToken)
	} else if len(rc.cookies) > 0 {
		for _, cookie := range rc.cookies {
			req.SetCookie(cookie)
		}
	}

	resp, err := req.Get(rc.serverURL + "/api/email/verify")
	if err != nil {
		return fmt.Errorf("email verification request failed: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("email verification failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	rc.log.Info("Email verification successful")
	return nil
}
