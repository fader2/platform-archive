package interfaces

import (
	"github.com/satori/go.uuid"
)

type MemDbManager struct {
}

func (m MemDbManager) FindFileByName(bucketName, fileName string, used DataUsed) (*File, error) {
	return nil, nil
}

func (m MemDbManager) FindFile(fileID uuid.UUID, used DataUsed) (*File, error) {
	return nil, nil
}

func (m MemDbManager) CreateFile(*File) error {
	return nil
}

func (m MemDbManager) CreateFileFrom(*File, DataUsed) error {
	return nil
}

func (m MemDbManager) UpdateFileFrom(*File, DataUsed) error {
	return nil
}

func (m MemDbManager) DeleteFile(fileID uuid.UUID) error {
	return nil
}

func (m MemDbManager) FindBucketByName(name string, used DataUsed) (*Bucket, error) {
	return nil, nil
}

func (m MemDbManager) FindBucket(bucketID uuid.UUID, used DataUsed) (*Bucket, error) {
	return nil, nil
}

func (m MemDbManager) CreateBucket(*Bucket) error {
	return nil
}

func (m MemDbManager) CreateBucketFrom(*Bucket, DataUsed) error {
	return nil
}

func (m MemDbManager) DeleteBucket(name string) error {
	return nil
}

func (m MemDbManager) UpdateBucket(*Bucket, DataUsed) error {
	return nil
}

func (m MemDbManager) EachBucket(func(*Bucket) error) error {
	return nil
}

func (m MemDbManager) EachFile(func(*File) error) error {
	return nil
}
