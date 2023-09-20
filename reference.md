# new post
```shell
cd ssg && hugo new announcements/moving_the_musings.md
git push
```

# docker build/run
```shell
docker build -f Dockerfile -t stormbuilder:local .
docker run -it --rm --env-file docker/odin.env stormbuilder:local
docker run --rm --env-file odin.env -v godin-cloud-build-sa.json:gcp-credentials.json odin-stormbuilder:local
```

# run sql migrations
```shell
python sqlmigrate.py
```