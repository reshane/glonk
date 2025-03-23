package store

import (
    "github.com/reshane/glonk/types"
)

type Store interface {
    // generic over DataType interface
    Get(types.MetaData, int64, int64) (types.DataType, error)
    GetByGuid(types.MetaData, string) (types.DataType, error)
    GetByQueries(types.MetaData, []types.Query, int64) ([]types.DataType, error)
    Create(types.DataType) (types.DataType, error)
    Update(types.DataType) (types.DataType, error)
    Delete(types.MetaData, int64, int64) (error)
}

