package encoding

func (e *Encoder) splitIntoShards(data []byte) ([][]byte, int) {
	shardSize := (len(data) + e.dataShards - 1) / e.dataShards
	if shardSize == 0 {
		shardSize = 1
	}

	totalShards := e.dataShards + e.parityShards
	shards := make([][]byte, totalShards)
	for i := range shards {
		shards[i] = make([]byte, shardSize)
	}

	for i := 0; i < len(data); i++ {
		shard := i / shardSize
		if shard >= e.dataShards {
			break
		}
		shards[shard][i%shardSize] = data[i]
	}

	return shards, shardSize
}

func (e *Encoder) joinShards(shards [][]byte, shardSize int) []byte {
	totalShards := e.dataShards + e.parityShards
	result := make([]byte, shardSize*totalShards)

	for i, shard := range shards {
		copy(result[i*shardSize:], shard)
	}

	return result
}
