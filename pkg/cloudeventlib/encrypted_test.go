// Package msg provides functionality for creating and parsing encrypted CloudEvents.
package msg

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var payload = `[{"id":1,"severity":"critical","tag":"alarm_tpellet"},{"id":2,"severity":"critical","tag":"alarm_twater"},{"id":3,"severity":"critical","tag":"over_tboard"},{"id":4,"severity":"error","tag":"hwfail_in_tzone1"},{"id":5,"severity":"error","tag":"hwfail_in_tzone2"},{"id":7,"severity":"critical","tag":"hwfail_in_tfumes"},{"id":9,"severity":"error","tag":"hwfail_in_twater"},{"id":15,"severity":"error","tag":"hwfail_in_pwater"},{"id":16,"severity":"error","tag":"hwfail_out_vfumes"},{"id":17,"severity":"error","tag":"hwfail_out_vauger"},{"id":18,"severity":"error","tag":"hwfail_out_pump"},{"id":19,"severity":"critical","tag":"over_tfumes"},{"id":20,"severity":"error","tag":"under_pwater"},{"id":21,"severity":"error","tag":"over_pwater"},{"id":23,"severity":"error","tag":"err_flame_start1"},{"id":24,"severity":"error","tag":"err_flame_start2"},{"id":25,"severity":"error","tag":"err_flame_steady"},{"id":26,"severity":"error","tag":"err_flame_work"},{"id":27,"severity":"error","tag":"err_pressure"},{"id":28,"severity":"error","tag":"err_pellet_door_open"}]`

// TestCreateCloudEvent tests the creation of a CloudEvent with encrypted payload.
// It verifies that:
// - The CloudEvent is created successfully
// - The payload is encrypted correctly
// - The CloudEvent can be parsed and decrypted successfully
func TestCreateCloudEvent(t *testing.T) {
	num := 1
	privkeys := make([]*ecdsa.PrivateKey, num)
	pubkeys := make([]*ecdsa.PublicKey, num)
	recipients := make([]CloudEventRecipient, num)

	for i := range privkeys {
		var err error
		privkeys[i], err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatal(err.Error())
		}
		pubkeys[i] = &privkeys[i].PublicKey

		recipients[i] = CloudEventRecipient{
			ID:     fmt.Sprintf("tenant-%d", i),
			PubKey: pubkeys[i],
		}
	}

	// Assuming CreateCloudEvent takes some parameters and returns an event and an error
	event, err := CreateCloudEvent(CloudEventOptions{
		Compression: CompressionBrotli,
		Payload:     []byte(payload),
		Recipients:  recipients,
		ID:          "id",
		Subject:     "subject",
		Source:      "source",
		Type:        "type",
		Time:        time.Now(),
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	j, err := event.MarshalJSON()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	fmt.Printf("el: %v\n", len(j))
	fmt.Printf("j: %s\n", j)

	ce, err := ParseCloudEvent(event, []interface{}{privkeys[0]})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	pl := string(ce.Payload)

	if pl != payload {
		t.Fatalf("expected examplePayload, got %s", pl)
	}
}

// TestCloudEventOptionsValidate tests the validation functionality of CloudEventOptions.
// It verifies that:
// - A properly configured CloudEventOptions passes validation
// - Required fields are properly enforced
// - Empty or invalid options fail validation
func TestCloudEventOptionsValidate(t *testing.T) {
	// Generate test key
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	tests := []struct {
		name    string
		options CloudEventOptions
		wantErr bool
	}{
		{
			name: "valid options",
			options: CloudEventOptions{
				Compression: CompressionGzip,
				Payload:     []byte("test payload"),
				Recipients: []CloudEventRecipient{
					{
						ID:     "test-id",
						PubKey: &privKey.PublicKey,
					},
				},
				ID:      "test-id",
				Subject: "test-subject",
				Source:  "test-source",
				Type:    "test-type",
				Time:    time.Now(),
			},
			wantErr: false,
		},
		{
			name:    "empty options",
			options: CloudEventOptions{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.options.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestCreateAndParseCloudEvent performs an end-to-end test of the CloudEvent
// encryption and decryption process. It tests:
// - Creating a CloudEvent with encrypted payload
// - Verifying all CloudEvent fields and extensions
// - Parsing and decrypting the CloudEvent
// - Verifying the decrypted data matches the original
func TestCreateAndParseCloudEvent(t *testing.T) {
	// Generate test keys
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	// Test data
	testTime := time.Now()
	testPayload := []byte("test payload")

	options := CloudEventOptions{
		Compression: CompressionNone,
		Payload:     testPayload,
		Recipients: []CloudEventRecipient{
			{
				ID:     "test-id",
				PubKey: &privKey.PublicKey,
			},
		},
		ID:      "test-id",
		Subject: "test-subject",
		Source:  "test-source",
		Type:    "test-type",
		Time:    testTime,
	}

	// Create cloud event
	cloudEvent, err := CreateCloudEvent(options)
	require.NoError(t, err)
	require.NotNil(t, cloudEvent)

	// Verify cloud event fields
	assert.Equal(t, options.ID, cloudEvent.ID())
	assert.Equal(t, options.Source, cloudEvent.Source())
	assert.Equal(t, options.Subject, cloudEvent.Subject())
	assert.Equal(t, options.Type, cloudEvent.Type())
	assert.Equal(t, options.Time.UTC(), cloudEvent.Time().UTC())

	// Verify extensions
	extensions := cloudEvent.Extensions()
	compression, ok := extensions["compression"].(string)
	require.True(t, ok)
	assert.Equal(t, string(options.Compression), compression)

	recipients, ok := extensions["recipients"].(string)
	require.True(t, ok)
	assert.Equal(t, options.Recipients[0].ID, recipients)

	// Parse cloud event
	parsedOptions, err := ParseCloudEvent(cloudEvent, []interface{}{privKey})
	require.NoError(t, err)
	require.NotNil(t, parsedOptions)

	// Verify parsed options match original options
	assert.Equal(t, options.ID, parsedOptions.ID)
	assert.Equal(t, options.Source, parsedOptions.Source)
	assert.Equal(t, options.Subject, parsedOptions.Subject)
	assert.Equal(t, options.Type, parsedOptions.Type)
	assert.Equal(t, options.Time.UTC(), parsedOptions.Time.UTC())
	assert.Equal(t, options.Compression, parsedOptions.Compression)
	assert.Equal(t, len(options.Recipients), len(parsedOptions.Recipients))
	assert.Equal(t, options.Recipients[0].ID, parsedOptions.Recipients[0].ID)
	assert.Equal(t, string(options.Payload), string(parsedOptions.Payload))
}

// TestParseCloudEventErrors verifies error handling in ParseCloudEvent.
// It tests various error conditions including:
// - Handling of nil CloudEvent
// - Missing required extensions
// - Invalid or missing compression settings
// - Invalid or missing recipient information
func TestParseCloudEventErrors(t *testing.T) {
	// Generate test keys
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	tests := []struct {
		name      string
		event     *event.Event
		keys      []interface{}
		wantError string
	}{
		{
			name:      "nil event",
			event:     nil,
			keys:      []interface{}{privKey},
			wantError: "cloud event is nil",
		},
		{
			name: "missing compression",
			event: func() *event.Event {
				e := event.New()
				e.SetID("test-id")
				return &e
			}(),
			keys:      []interface{}{privKey},
			wantError: "invalid compression extension type",
		},
		{
			name: "missing recipients",
			event: func() *event.Event {
				e := event.New()
				e.SetID("test-id")
				e.SetExtension("compression", string(CompressionNone))
				return &e
			}(),
			keys:      []interface{}{privKey},
			wantError: "invalid recipients extension type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseCloudEvent(tt.event, tt.keys)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantError)
		})
	}
}
