## ProvectusInternship_TiahurDmytro


### Setup
Solution can be run using ```docker-compose.yml``` from the root of the repo.
```
docker-compose up
```
Solution is written in Golang using [github.com/minio/minio-go](https://github.com/minio/minio-go) for working with Minio.
Result is prebuilt and used from the docker container ```dmytrothr/user-aggregator```. ENV parameters are the following (can be changed in ```docker-compose.yml```):
```
    environment:
      HTTP_SERVER_PORT: ':8080'
      MINIO_ROOT_PASSWORD: 'password'
      MINIO_ROOT_USER: 'admin'
      BUCKET_SERVICE_ADDR: 'minio:9000'
      REFRESH_PERIOD_MIN: '2'
```
If you'd like to change an HTTP port, you should also change it in ```ports``` section of ```user-aggregator``` service parameters:
```
    ports:
      - 8080:8080
```

#### Project source code
[03-solution]()

### Result will be represented as a response to:

```GET /data``` - endpoint with possible filters ```min_age=N```, ```max_age=N```, ```is_image_exists=True/False```
Eg:
```json
{
	"1000": {
		"ID": "1000",
		"Firstname": "Susan",
		"Lastname": "Lee",
		"Birthday": "1989-05-27T20:00:00Z",
		"PicturePath": "1000.png"
	},
	"1001": {
		"ID": "1001",
		"Firstname": "Rosa",
		"Lastname": "Garcia",
		"Birthday": "1991-04-02T21:00:00Z",
		"PicturePath": "1001.png"
	}
}
```

```GET /stats``` - endpoint with possible filters ```min_age=N```, ```max_age=N```, ```is_image_exists=True/False```.
Calculates the average age of the filtered users. Eg:
```json
{
	"Age": 29.2
}
```

```POST /data``` - forces the system to reaggregate users. Without it results will be recalculated each ```REFRESH_PERIOD_MIN``` minutes.
```
curl -X POST http://localhost:8080/data
```
```json
{
	"Message": "Data is being updated, you can reGET it."
}
```

Also, an aggregated file will be created as a result. Saved as ```processed_data/out.csv``` in the solution app folder and also as ```out.csv``` in ```processed-data``` bucket in Minio.