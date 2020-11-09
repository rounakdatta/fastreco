## fastreco

Fastreco is a simple command line based item Recommender which uses Item-To-Item Collaborative Filtering at its core. That is, when given a collection of interaction between users and items, it helps in finding highly associated pairs of items (or _'If you liked A, you might like B'_).

Fastreco expects a CSV file as the training interaction data (consisting at least user_id and item_id columns) and produces a processed JSON as the recommendation data. It also maintains a status file for minimal caching. Currently the recommendation computations expects the required columns to be integer. However, abstractions to map the ids to actual values are in progress.

### How much accuracy / complexity is supported?
Currently this is a very simple implementation taking into account only interactions, and not contextual similarity. It uses simple statistical algorithms like [Log Likelihood](https://en.wikipedia.org/wiki/Likelihood_function).

### Why call it fast?
It is fast because it performs significantly fast than the pandas-based approach in Python, thanks to [qframe](https://github.com/tobgu/qframe)'s enhanced DataFrame processing as well as introduced concurrency & parallelism in this implementation. There are equivalent implementations in [Python](https://github.com/oni-on/item-collaborative-filtering), [Rust](https://github.com/sscdotopen/recoreco) and many more languages and we intend to publish detailed benchmark of performance. 

## Example Usage
We first need to prepare the binary `fastreco` using
```bash
go build
```

Next, lets say we want to experiment with the GoodReads books dataset,
```bash
# grabbing the training data
curl -O https://raw.githubusercontent.com/zygmuntz/goodbooks-10k/master/ratings.csv

# user_id (int) gives the user identifications
# book_id (int) gives the item identifications
# rating (int) gives a metric of whether the item is liked (>= 5) or not
head -n 3 ratings.csv
# user_id,book_id,rating
# 1,258,5
# 2,4081,4

# computing top 5 recommendations for item id 1212
./fastreco --input-file "ratings.csv" \
	--user-column "user_id" \
	--item-column "book_id" \
	--liked-column "rating" \
	--liked-threshold 5 \
	--item-id 1212
# [2 24 23 19 37 6 1 5 7 20]
```

### Force re-computation
By default fastreco will cache the recommendation results on per-user id basis. However, use of `--force` flag makes a fresh re-computation for that particular user.

### Computing recommendations for all users
Although a computationally costly operation, we can skip the `item-id` flag to demand processing of recommendation for each and every unique user.