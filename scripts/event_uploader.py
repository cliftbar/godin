import zipfile

from typing import List, Optional
from pathlib import PurePath, Path

from google.cloud import storage


def list_blobs_with_prefix(bucket_name, prefix, delimiter=None):
    """Lists all the blobs in the bucket that begin with the prefix.

    This can be used to list all blobs in a "folder", e.g. "public/".

    The delimiter argument can be used to restrict the results to only the
    "files" in the given "folder". Without the delimiter, the entire tree under
    the prefix is returned. For example, given these blobs:

        a/1.txt
        a/b/2.txt

    If you specify prefix ='a/', without a delimiter, you'll get back:

        a/1.txt
        a/b/2.txt

    However, if you specify prefix='a/' and delimiter='/', you'll get back
    only the file directly under 'a/':

        a/1.txt

    As part of the response, you'll also get back a blobs.prefixes entity
    that lists the "subfolders" under `a/`:

        a/b/
    """

    storage_client = storage.Client()

    # Note: Client.list_blobs requires at least package version 1.17.0.
    blobs = storage_client.list_blobs(bucket_name, prefix=prefix, delimiter=delimiter)

    blobs = [x for x in blobs]

    print("Blobs:")
    for blob in blobs:
        print(blob.name)

    if delimiter:
        print("Prefixes:")
        for prefix in blobs.prefixes:
            print(prefix)

    return blobs


def rename_blob(bucket_name, blob_name, new_name):
    """Renames a blob."""
    # The ID of your GCS bucket
    # bucket_name = "your-bucket-name"
    # The ID of the GCS object to rename
    # blob_name = "your-object-name"
    # The new ID of the GCS object
    # new_name = "new-object-name"

    storage_client = storage.Client()
    bucket = storage_client.bucket(bucket_name)
    blob = bucket.blob(blob_name)

    new_blob = bucket.rename_blob(blob, new_name)

    print("Blob {} has been renamed to {}".format(blob.name, new_blob.name))


def upload_blob(bucket_name, source_file_name, destination_blob_name):
    """Uploads a file to the google storage bucket."""

    storage_client = storage.Client()
    bucket = storage_client.bucket(bucket_name)
    blob = bucket.blob(destination_blob_name)

    blob.upload_from_filename(source_file_name)

    print(
        "File {} uploaded to Storage Bucket with Blob name  {} successfully .".format(
            source_file_name, destination_blob_name
        )
    )


def move_blob(bucket_name, blob_name, destination_bucket_name, destination_blob_name):
    """Moves a blob from one bucket to another with a new name."""
    # The ID of your GCS bucket
    # bucket_name = "your-bucket-name"
    # The ID of your GCS object
    # blob_name = "your-object-name"
    # The ID of the bucket to move the object to
    # destination_bucket_name = "destination-bucket-name"
    # The ID of your new GCS object (optional)
    # destination_blob_name = "destination-object-name"

    storage_client = storage.Client()

    source_bucket = storage_client.bucket(bucket_name)
    source_blob = source_bucket.blob(blob_name)
    destination_bucket = storage_client.bucket(destination_bucket_name)

    blob_copy = source_bucket.copy_blob(
        source_blob, destination_bucket, destination_blob_name
    )
    source_bucket.delete_blob(blob_name)

    print(
        "Blob {} in bucket {} moved to blob {} in bucket {}.".format(
            source_blob.name,
            source_bucket.name,
            blob_copy.name,
            destination_bucket.name,
        )
    )


def zipdir(path: str, ziph: zipfile.ZipFile):
    # ziph is zipfile handle
    # for root, dirs, files in os.walk(path):
    zip_path: Path = Path(path)
    for file in zip_path.glob("*"):
        fi_path: Path = Path(file)
        if fi_path.is_dir() or fi_path.name == ".DS_Store" or fi_path.suffix == ".zip":
            continue
        print(file)
        ziph.write(
            fi_path,
            fi_path.name
        )


def upload_event(storm_qgis_filename: Optional[str]):
    if storm_qgis_filename is None:
        print("Set a qgis filename")
        exit(2)
    bucket = "godin_hurricane_data"
    data_path: str = "./data"
    # storm_name = "sandy2012"
    # filename = f"{storm_name}_100x100_20210831T1700-04"

    fi_splits: List[str] = storm_qgis_filename.split("_")
    storm_name: str = fi_splits[0]
    filename, _ = storm_qgis_filename.split(".", maxsplit=1)

    # zip relavent files
    zipf = zipfile.ZipFile(f"{data_path}/{storm_name}/{filename}.zip", 'w', zipfile.ZIP_DEFLATED)
    zipdir(f"{data_path}/{storm_name}", zipf)
    zipf.close()

    blobs = list_blobs_with_prefix(bucket_name=bucket, prefix=f"{storm_name}/latest")
    for blob in blobs:
        blob_path: PurePath = PurePath(blob.name)
        if blob.name[-1] == "/":
            continue
        move_blob(bucket_name=bucket, blob_name=blob.name, destination_bucket_name=bucket, destination_blob_name=f"{storm_name}/past/{blob_path.name}")

    # ZIP data
    source = f"{data_path}/{storm_name}/{filename}.zip"
    dest = f"{storm_name}/latest/{filename}.zip"
    upload_blob(bucket_name=bucket, source_file_name=source, destination_blob_name=dest)

    # gis
    source = f"{data_path}/{storm_name}/{filename}.png"
    dest = f"{storm_name}/latest/{filename}.png"
    upload_blob(bucket_name=bucket, source_file_name=source, destination_blob_name=dest)


if __name__ == "__main__":
    qgis_filename: str = "matthew2016_100x100_2021-09-07T14:37:00+00:00.png"
    upload_event(qgis_filename)
