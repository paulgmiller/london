package backend

import (
	"fmt"
	"testing"

	"github.com/khenidak/london/pkg/backend/consts"
	storageerrors "github.com/khenidak/london/pkg/backend/storageerrors"
	basictestutils "github.com/khenidak/london/test/utils/basic"
)

func TestUpdate(t *testing.T) {
	c := basictestutils.MakeTestConfig(t)

	be, err := NewBackend(c)
	if err != nil {
		t.Fatalf("failed to create backend with err:%v", err)
	}

	key := fmt.Sprintf("/%s/%s/%s", randStringRunes(8), randStringRunes(8), randStringRunes(8))
	val := randStringRunes(1024 * 1024)

	insertRev, err := be.Insert(key, []byte(val), 1)
	if err != nil {
		t.Fatalf("failed to insert with err :%v", err)
	}

	// trying to update an nonexistent record should yeild into 404
	_, err = be.Update("idontexist", []byte("someval"), 79, 1)
	if err == nil || !storageerrors.IsNotFoundError(err) {
		t.Fatalf("expected a not found error instead got:%v", err)
	}

	// trying to update an incorrect rev should yeild into err conflict and
	// existing record returned
	currentRecord, err := be.Update(key, []byte("newVal"), 90909090, 1)
	if err == nil || !storageerrors.IsConflictError(err) {
		t.Fatalf("expected a conflict error instead got:%v", err)
	}

	if string(currentRecord.Value()) != string(val) {
		t.Fatalf("expected that current record returned from failed update to have equal value to the one originally inserted")
	}

	// perform a valid update
	updatedVal := []byte(randStringRunes(1024 * 1024 * 10))[:consts.DataRowMaxSize*3]
	currentRecord, err = be.Update(key, updatedVal, insertRev, 1)
	if err != nil {
		t.Fatalf("unexpected error during a valid update:%v", err)
	}

	if string(currentRecord.Value()) != string(updatedVal) {
		t.Fatalf("expected record returned from valid update to carry the updated value")
	}

	if currentRecord.ModRevision() == currentRecord.CreateRevision() {
		t.Fatalf("expected modRev != createRev %v==%v", currentRecord.ModRevision(), currentRecord.CreateRevision())
	}

	if currentRecord.CreateRevision() != insertRev {
		t.Fatalf("expected prevRev == insertedRev %v!=%v", currentRecord.PrevRevision(), insertRev)
	}

	// get to ensure update actually wrote the data
	updatedRecord, _, err := be.Get(key, currentRecord.ModRevision())
	if err != nil {
		t.Fatalf("failed to get after update with err :%v", err)
	}

	if string(updatedRecord.Value()) != string(updatedVal) {
		t.Fatalf("data after get does not match updated data")
	}

	// update again to ensure that events insertion does not conflict
	_, err = be.Update(key, updatedVal, currentRecord.ModRevision(), 1)
	if err != nil {
		t.Fatalf("failed to repeat valid update %v", err)
	}

}
