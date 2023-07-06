package eth

import (
	"ethernal/explorer/db"
	"sync"
)

type nftMetadataDictionary struct {
	lock      sync.RWMutex
	items     map[string]bool
	itemsData chan itemsData
}

type itemsData struct {
	metadata   []*db.NftMetadata
	attributes []*db.NftMetadataAttribute
}

var lockDictionary = &sync.Mutex{}

var dictionaryInstance *nftMetadataDictionary

// create a singleton instance of a nft metadata dictionary
func GetMetadataDictionaryInstance() *nftMetadataDictionary {
	if dictionaryInstance == nil {
		lockDictionary.Lock()
		defer lockDictionary.Unlock()
		if dictionaryInstance == nil {
			dictionaryInstance = &nftMetadataDictionary{
				items:     make(map[string]bool),
				itemsData: make(chan itemsData),
			}
		}
	}
	return dictionaryInstance
}

// TryAdd method adds a metadata to the dictionary, if it does not already exist
func (dict *nftMetadataDictionary) TryAdd(key string, value bool) bool {
	dict.lock.Lock()
	defer dict.lock.Unlock()
	_, ok := dict.items[key]
	if !ok {
		dict.items[key] = value
		return true
	}
	return false
}

// TryRemove removes a metadata from the dictionary, if it exists
func (dict *nftMetadataDictionary) TryRemove(key string) bool {
	dict.lock.Lock()
	defer dict.lock.Unlock()
	_, ok := dict.items[key]
	if ok {
		delete(dict.items, key)
	}
	return ok
}

// TryRemoveRange removes a metadata range from the dictionary, if it exists
func (dict *nftMetadataDictionary) TryRemoveRange(keys []string) {
	dict.lock.Lock()
	defer dict.lock.Unlock()
	for _, key := range keys {
		_, ok := dict.items[key]
		if ok {
			delete(dict.items, key)
		}
	}
}
