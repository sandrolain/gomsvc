// Package msg provides functionality for creating and parsing encrypted CloudEvents.
// It supports various compression methods and encryption using JWE (JSON Web Encryption).
// The package is designed to work with the CloudEvents specification while adding
// secure message passing capabilities through encryption and recipient management.
package msg

import (
	"fmt"
	"strings"
	"time"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/go-playground/validator/v10"
	"github.com/sandrolain/gomsvc/pkg/jwxlib"
	"github.com/sandrolain/gomsvc/pkg/ziplib"
)

// CompressionMethod represents the type of compression used for the CloudEvent payload.
type CompressionMethod string

// CloudEventRecipient represents a recipient of an encrypted CloudEvent.
// Each recipient has an ID and a public key used for encryption.
type CloudEventRecipient struct {
	// ID is a unique identifier for the recipient
	ID string `validate:"required"`
	// PubKey is the recipient's public key used for encryption
	PubKey interface{} `validate:"required"`
}

// CloudEventOptions contains all the necessary information to create or parse
// an encrypted CloudEvent. It includes compression settings, payload data,
// recipient information, and standard CloudEvent fields.
type CloudEventOptions struct {
	// Compression specifies the compression method used for the payload
	Compression CompressionMethod `validate:"required"`
	// Payload is the raw data to be encrypted and included in the CloudEvent
	Payload []byte `validate:"required"`
	// Recipients is a list of recipients who can decrypt the CloudEvent
	Recipients []CloudEventRecipient `validate:"required,dive,required"`
	// ID is the unique identifier of the CloudEvent
	ID string `validate:"required"`
	// Subject is the subject of the CloudEvent
	Subject string `validate:"required"`
	// Source identifies the context in which an event happened
	Source string `validate:"required"`
	// Type represents the type of event related to the originating occurrence
	Type string `validate:"required"`
	// Time is the timestamp of when the occurrence happened
	Time time.Time `validate:"required"`
}

const (
	// CompressionNone indicates no compression is used
	CompressionNone CompressionMethod = "none"
	// CompressionGzip indicates Gzip compression is used
	CompressionGzip CompressionMethod = "gzip"
	// CompressionBrotli indicates Brotli compression is used
	CompressionBrotli CompressionMethod = "brotli"
)

func (o *CloudEventOptions) Validate() error {
	validate := validator.New()
	return validate.Struct(o)
}

// CreateCloudEvent creates a new encrypted CloudEvent using the provided options.
// The function performs the following steps:
// 1. Validates the input options
// 2. Compresses the payload using the specified compression method
// 3. Encrypts the compressed payload using the recipients' public keys
// 4. Creates a CloudEvent with the encrypted data and necessary extensions
func CreateCloudEvent(options CloudEventOptions) (*event.Event, error) {
	if err := options.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	var compressedPayload []byte
	var err error

	switch options.Compression {
	case CompressionGzip:
		compressedPayload, err = ziplib.GzipCompress(options.Payload)
	case CompressionBrotli:
		compressedPayload, err = ziplib.BrotliCompress(options.Payload)
	case CompressionNone:
		compressedPayload = options.Payload
	default:
		return nil, fmt.Errorf("unsupported compression method: %s", options.Compression)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to compress payload: %w", err)
	}

	// Prepare public keys for encryption
	pubKeys := make([]interface{}, len(options.Recipients))
	recipientsIds := make([]string, len(options.Recipients))

	for i, rec := range options.Recipients {
		pubKeys[i] = rec.PubKey
		recipientsIds[i] = rec.ID
	}

	// Encrypt the compressed payload using jwxlib
	encryptedPayload, err := jwxlib.JweEncrypt(compressedPayload, pubKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt payload: %w", err)
	}

	// Create cloud event
	cloudEvent := event.New()
	cloudEvent.SetID(options.ID)
	cloudEvent.SetSource(options.Source)
	cloudEvent.SetSubject(options.Subject)
	cloudEvent.SetType(options.Type)
	cloudEvent.SetTime(options.Time)
	cloudEvent.SetExtension("compression", string(options.Compression))
	cloudEvent.SetExtension("recipients", strings.Join(recipientsIds, ","))

	if err := cloudEvent.SetData(event.ApplicationJSON, encryptedPayload); err != nil {
		return nil, fmt.Errorf("failed to set cloud event data: %w", err)
	}

	return &cloudEvent, nil
}

// ParseCloudEvent parses an encrypted CloudEvent and returns the decrypted options.
// The function performs the following steps:
// 1. Extracts compression and recipient information from CloudEvent extensions
// 2. Decrypts the data using the provided decryption keys
// 3. Decompresses the decrypted data
// 4. Returns the original CloudEvent options with decrypted payload
//
// The keys parameter should contain the private keys corresponding to one of the
// recipients specified in the CloudEvent.
func ParseCloudEvent(cloudEvent *event.Event, keys []interface{}) (*CloudEventOptions, error) {
	if cloudEvent == nil {
		return nil, fmt.Errorf("cloud event is nil")
	}

	// Get compression method
	compressionExt := cloudEvent.Extensions()["compression"]
	compression, ok := compressionExt.(string)
	if !ok {
		return nil, fmt.Errorf("invalid compression extension type")
	}

	// Get recipients
	recipientsExt := cloudEvent.Extensions()["recipients"]
	recipientsStr, ok := recipientsExt.(string)
	if !ok {
		return nil, fmt.Errorf("invalid recipients extension type")
	}
	recipientIds := strings.Split(recipientsStr, ",")

	// Get encrypted data
	encryptedData := cloudEvent.Data()

	// Decrypt data using jwxlib
	decryptedData, err := jwxlib.JweDecrypt(encryptedData, keys)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt payload: %w", err)
	}

	// Decompress data
	var decompressedData []byte
	switch CompressionMethod(compression) {
	case CompressionGzip:
		decompressedData, err = ziplib.GzipDecompress(decryptedData)
	case CompressionBrotli:
		decompressedData, err = ziplib.BrotliDecompress(decryptedData)
	case CompressionNone:
		decompressedData = decryptedData
	default:
		return nil, fmt.Errorf("unsupported compression method: %s", compression)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to decompress payload: %w", err)
	}

	// Create recipients list
	recipients := make([]CloudEventRecipient, len(recipientIds))
	for i, id := range recipientIds {
		recipients[i] = CloudEventRecipient{
			ID: id,
		}
	}

	return &CloudEventOptions{
		ID:          cloudEvent.ID(),
		Source:      cloudEvent.Source(),
		Subject:     cloudEvent.Subject(),
		Type:        cloudEvent.Type(),
		Time:        cloudEvent.Time(),
		Compression: CompressionMethod(compression),
		Payload:     decompressedData,
		Recipients:  recipients,
	}, nil
}
