//go:build rethdb

package sources

import (
	"encoding/json"
	"fmt"

	"unsafe"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

/*
#cgo LDFLAGS: -L../rethdb-reader/target/release -lrethdbreader
#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>

typedef struct {
    char* data;
    size_t data_len;
    bool error;
} ReceiptsResult;

extern ReceiptsResult* rdb_read_receipts(const uint8_t* block_hash, size_t block_hash_len, const char* db_path);
extern void rdb_free_receipts(ReceiptsResult* receipts_result);
extern char* rdb_get_receipts_data(ReceiptsResult* receipts_result);
extern size_t rdb_get_receipts_data_len(ReceiptsResult* receipts_result);
extern char* rdb_get_receipts_error(ReceiptsResult* receipts_result);
*/
import "C"

// FetchRethReceipts fetches the receipts for the given block hash directly from the Reth Database
// and populates the given results slice pointer with the receipts that were found.
func FetchRethReceipts(dbPath string, blockHash *common.Hash) (types.Receipts, error) {
	if blockHash == nil {
		return nil, fmt.Errorf("Must provide a block hash to fetch receipts for.")
	}

	// Convert the block hash to a C byte array and defer its deallocation
	cBlockHash := C.CBytes(blockHash[:])
	defer C.free(cBlockHash)

	// Convert the db path to a C string and defer its deallocation
	cDbPath := C.CString(dbPath)
	defer C.free(unsafe.Pointer(cDbPath))

	// Call the C function to fetch the receipts from the Reth Database
	receiptsResult := C.rdb_read_receipts((*C.uint8_t)(cBlockHash), C.size_t(len(blockHash)), cDbPath)
	if receiptsResult == nil {
		return nil, fmt.Errorf("unexpected null Receipts Result")
	}
	// Free the memory allocated by the C code
	defer C.rdb_free_receipts(receiptsResult)

	receiptsResultErr := C.rdb_get_receipts_error(receiptsResult)
	if receiptsResultErr != nil {
		return nil, fmt.Errorf("Error fetching receipts from Reth Database: %s", C.GoString(receiptsResultErr))
	}

	receiptsData := C.rdb_get_receipts_data(receiptsResult)
	receiptsDataLen := C.int(C.rdb_get_receipts_data_len(receiptsResult))

	// Convert the returned JSON string to Go string and parse it
	receiptsJSON := C.GoStringN(receiptsData, receiptsDataLen)
	var receipts types.Receipts
	if err := json.Unmarshal([]byte(receiptsJSON), &receipts); err != nil {
		return nil, err
	}

	return receipts, nil
}