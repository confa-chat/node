package uuid

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	fuuid "github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

var Nil = UUID{}

type UUIDList = []UUID

type UUID struct {
	fuuid.UUID
}

func New() UUID {
	return UUID{fuuid.Must(fuuid.NewV7())}
}

func NewFromTime(t time.Time) UUID {
	gen := fuuid.NewGenWithOptions(
		fuuid.WithEpochFunc(func() time.Time { return t }),
	)
	return UUID{fuuid.Must(gen.NewV7())}
}

func NewP() *UUID {
	return &UUID{fuuid.Must(fuuid.NewV7())}
}

func FromString(text string) (UUID, error) {
	u, err := fuuid.FromString(text)
	if err != nil {
		return Nil, err
	}

	return UUID{u}, nil
}

func MustFromString(text string) UUID {
	u, err := fuuid.FromString(text)
	if err != nil {
		panic(err)
	}

	return UUID{u}
}

func FromBytes(input []byte) (UUID, error) {
	u, err := fuuid.FromBytes(input)
	if err != nil {
		return Nil, err
	}

	return UUID{u}, nil
}

func (a *UUID) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	if s == "" {
		a.UUID = fuuid.Nil
		return nil
	}

	return a.UUID.Parse(s)
}

func (a UUID) MarshalJSON() ([]byte, error) {
	if a.IsNil() {
		return json.Marshal("")
	}

	return json.Marshal(a.UUID)
}

// UnmarshalGQL implements the graphql.Unmarshaler interface
func (u *UUID) UnmarshalGQL(v interface{}) error {
	id, ok := v.(string)
	if !ok {
		return fmt.Errorf("uuid must be a string")
	}

	return u.Parse(id)
}

// MarshalGQL implements the graphql.Marshaler interface
func (u UUID) MarshalGQL(w io.Writer) {
	b := []byte(strconv.Quote(u.String()))
	_, err := w.Write(b)
	if err != nil {
		panic(err)
	}
}

const uuidSubtype = 4

// MarshalBSONValue для официального mongo драйвера (go.mongodb.org/mongo-driver/mongo)
func (id UUID) MarshalBSONValue() (bsontype.Type, []byte, error) {
	if id.IsNil() {
		return bsontype.Null, nil, nil
	} else {
		bin := bsoncore.AppendBinary(nil, uuidSubtype, id.UUID.Bytes())
		return bson.TypeBinary, bin, nil
	}
}

// MarshalBSONValue для официального mongo драйвера (go.mongodb.org/mongo-driver/mongo)
func (id *UUID) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	switch t {
	case bsontype.Null, bsontype.Undefined:
		id = &UUID{} // nil uuid value 00000000-0000-0000-0000-000000000000
		return nil
	case bsontype.Binary:
		subtype, bin, _, ok := bsoncore.ReadBinary(data)
		if !ok {
			return errors.New("invalid bson binary value")
		}
		if subtype != uuidSubtype && subtype != 3 { // 3 is a deprecated uuid subtype, used for compatibility
			return fmt.Errorf("unsupported binary subtype for uuid: %d", subtype)
		}
		return id.UUID.UnmarshalBinary(bin)
	default:
		return fmt.Errorf("unsupported value type for uuid: %s", t.String())
	}
}

var _ pgtype.UUIDValuer = UUID{}

// UUIDValue implements pgtype.UUIDValuer.
func (a UUID) UUIDValue() (pgtype.UUID, error) {
	return pgtype.UUID{
		Bytes: a.UUID,
		Valid: true,
	}, nil
}

var _ pgtype.UUIDScanner = (*UUID)(nil)

// ScanUUID implements pgtype.UUIDScanner.
func (a *UUID) ScanUUID(v pgtype.UUID) error {
	if !v.Valid {
		return fmt.Errorf("cannot scan NULL into *uuid.UUID")
	}

	a.UUID = v.Bytes
	return nil
}
