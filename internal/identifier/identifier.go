package identifier

import (
	"errors"
	"math/rand"
	"regexp"
	"sync"
	"sync/atomic"
	"time"
)

const (
	prefix            string = "c-"
	alphaNumChars     string = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	timeSeqLen        int    = 8
	randSeqLen        int    = 6
	counterSeqLen     int    = 4
	identifierLength  int    = 20
	validationPattern string = "^c-[0-9A-Za-z]{18}$"
)

type Identifier struct {
	bytes [identifierLength]byte
}

func (i Identifier) String() string {
	return string(i.bytes[:])
}

func (i Identifier) Bytes() [identifierLength]byte {
	return i.bytes
}

func (i *Identifier) Scan(value interface{}) error {
	if value == nil {
		return errors.New("nil value")
	}

	switch v := value.(type) {
	case string:
		id, err := FromString(v)
		if err != nil {
			return err
		}
		*i = id
	case []byte:
		id, err := FromBytes(v)
		if err != nil {
			return err
		}
		*i = id
	default:
		return errors.New("unsupported data type")
	}

	return nil
}

var (
	counter                uint32
	rngPool                sync.Pool
	regexValidationPattern *regexp.Regexp
)

// Precompiling regex and initializing the RNG pool for performance gains
func init() {
	regexValidationPattern = regexp.MustCompile(validationPattern)

	rngPool = sync.Pool{
		New: func() interface{} {
			return rand.New(rand.NewSource(time.Now().UnixNano()))
		},
	}
}

func New() (Identifier, error) {
	count := atomic.AddUint32(&counter, 1)

	prefixLen := len(prefix)
	timeSeq := getTimeSequence()
	counterSeq := getCounterSequence(int(count))
	randSeq := getRandSequence()

	buff := [identifierLength]byte{}
	copy(buff[0:], prefix)
	copy(buff[prefixLen:], timeSeq[:])
	copy(buff[prefixLen+len(timeSeq):], counterSeq[:])
	copy(buff[prefixLen+len(timeSeq)+len(counterSeq):], randSeq[:])

	return Identifier{bytes: buff}, nil
}

func FromString(str string) (Identifier, error) {
	byteStr := []byte(str)
	if !validateLength(byteStr) {
		return Identifier{}, ValidationError{reason: BadLength}
	}

	if !validatePattern(byteStr) {
		return Identifier{}, ValidationError{reason: BadFormat}
	}

	buff := [identifierLength]byte{}
	copy(buff[:], byteStr)

	return Identifier{bytes: buff}, nil
}

func FromBytes(b []byte) (Identifier, error) {
	if !validateLength(b) {
		return Identifier{}, ValidationError{reason: BadLength}
	}

	if !validatePattern(b) {
		return Identifier{}, ValidationError{reason: BadFormat}
	}

	buff := [identifierLength]byte{}
	copy(buff[:], b)

	return Identifier{bytes: buff}, nil
}

func Validate(id Identifier) (bool, error) {
	byteSlice := id.bytes[:]
	valid := validateLength(byteSlice) && validatePattern(byteSlice)

	return valid, nil
}

func getTimeSequence() [timeSeqLen]byte {
	timeInt := int(time.Now().UnixMilli())
	encodedBuffer := [timeSeqLen]byte{}

	encodeToAlphaNums(timeInt, encodedBuffer[:])

	return encodedBuffer
}

func getCounterSequence(count int) [counterSeqLen]byte {
	counterBuffer := [counterSeqLen]byte{}
	encodeToAlphaNums(count, counterBuffer[:])

	return counterBuffer
}

func getRandSequence() [randSeqLen]byte {
	rng := rngPool.Get().(*rand.Rand)
	defer rngPool.Put(rng)

	var result [randSeqLen]byte
	for i := 0; i < len(result); i++ {
		randomNum := rng.Intn(len(alphaNumChars))
		result[i] = alphaNumChars[randomNum]
	}

	return result
}

func encodeToAlphaNums(num int, buff []byte) string {
	encodingBase := len(alphaNumChars)

	for i := len(buff) - 1; i >= 0; i-- {
		remainder := num % encodingBase
		val := alphaNumChars[remainder]
		buff[i] = val

		num = num / encodingBase
	}

	return string(buff[:])
}

func validateLength(b []byte) bool {
	return len(b) == identifierLength
}

func validatePattern(b []byte) bool {
	return regexValidationPattern.Match(b)
}
