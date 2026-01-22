package lib

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// TranscriptSegment represents a segment of transcribed text
type TranscriptSegment struct {
	Text      string
	StartTime int64 // milliseconds
	EndTime   int64 // milliseconds
}

// LoadWAVAsFloat32 loads a WAV file and returns float32 samples
func LoadWAVAsFloat32(filepath string) ([]float32, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open WAV file: %w", err)
	}
	defer file.Close()

	// Skip WAV header (44 bytes for basic WAV)
	// This is a simplified WAV parser that assumes 16kHz mono 16-bit PCM
	_, err = file.Seek(44, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to seek past WAV header: %w", err)
	}

	// Read the rest as 16-bit samples
	var samples []float32
	for {
		var sample int16
		err := binary.Read(file, binary.LittleEndian, &sample)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read sample: %w", err)
		}

		// Convert 16-bit int to float32 in range [-1.0, 1.0]
		samples = append(samples, float32(sample)/32768.0)
	}

	return samples, nil
}
