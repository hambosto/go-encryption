package encoding

func (r *ReedSolomon) splitIntoShards(data []byte) ([][]byte, int) {
	shardSize := (len(data) + r.dataShards - 1) / r.dataShards
	if shardSize == 0 {
		shardSize = 1
	}

	totalShards := r.dataShards + r.parityShards
	shards := make([][]byte, totalShards)
	for i := range shards {
		shards[i] = make([]byte, shardSize)
	}

	for i := range data {
		shard := i / shardSize
		if shard >= r.dataShards {
			break
		}
		shards[shard][i%shardSize] = data[i]
	}

	return shards, shardSize
}

func (r *ReedSolomon) joinShards(shards [][]byte, shardSize int) []byte {
	totalShards := r.dataShards + r.parityShards
	result := make([]byte, shardSize*totalShards)

	for i, shard := range shards {
		copy(result[i*shardSize:], shard)
	}

	return result
}
