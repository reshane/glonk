package store

import (
    "github.com/reshane/glonk/types"
)

type Store interface {
    // generic over DataType interface
    Get(string, int64) (types.DataType, error)
    GetByGuid(string, string) (types.DataType, error)
    GetByQueries(string, []types.Query) ([]types.DataType, error)
    Create(types.DataType) (types.DataType, error)
    Update(types.DataType) (types.DataType, error)
    Delete(string, int64) (error)
}

