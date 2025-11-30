import zipfile
from typing import List, Optional
from pathlib import PurePath, Path
from types import SimpleNamespace

import boto3
from botocore.exceptions import ClientError
from mypy_boto3_s3.client import S3Client


def _s3_client(
    aws_access_key_id: Optional[str] = None,
    aws_secret_access_key: Optional[str] = None,
    region_name: Optional[str] = None,
    endpoint_url: Optional[str] = None,
) -> S3Client:
    # Pass through provided credentials/region directly; boto3 will ignore None and use its default chain.
    return boto3.client(
        "s3",
        aws_access_key_id=aws_access_key_id,
        aws_secret_access_key=aws_secret_access_key,
        region_name=region_name,
        endpoint_url=endpoint_url,
    )


def list_blobs_with_prefix(
    s3: S3Client,
    bucket_name: str,
    prefix: str,
    delimiter: Optional[str] = None,
):
    """Lists all the objects in the S3 bucket that begin with the prefix.

    Mirrors the interface of the GCS helper by returning items with a `.name` attribute.
    """
    paginator = s3.get_paginator("list_objects_v2")

    pagination_config = {"PageSize": 1000}
    operation_parameters = {
        "Bucket": bucket_name,
        "Prefix": prefix,
    }
    if delimiter:
        operation_parameters["Delimiter"] = delimiter

    results: List[SimpleNamespace] = []

    for page in paginator.paginate(**operation_parameters, PaginationConfig=pagination_config):
        contents = page.get("Contents", [])
        for obj in contents:
            # Wrap each object so callers can use `.name` like GCS blob
            results.append(SimpleNamespace(name=obj["Key"], size=obj.get("Size")))

    return results


def rename_blob(
    s3: S3Client,
    bucket_name: str,
    blob_name: str,
    new_name: str,
):
    """Renames an object by copying then deleting the original.
    Equivalent to a move within the same bucket.
    """
    copy_source = {"Bucket": bucket_name, "Key": blob_name}
    try:
        s3.copy_object(Bucket=bucket_name, CopySource=copy_source, Key=new_name)
        s3.delete_object(Bucket=bucket_name, Key=blob_name)
    except ClientError as e:
        raise RuntimeError(f"Failed to rename {blob_name} to {new_name} in {bucket_name}: {e}")


def upload_blob(
    s3: S3Client,
    bucket_name: str,
    source_file_name: str,
    destination_blob_name: str,
):
    """Uploads a file to the S3 bucket."""
    try:
        s3.upload_file(source_file_name, bucket_name, destination_blob_name)
    except ClientError as e:
        raise RuntimeError(f"Failed to upload {source_file_name} to s3://{bucket_name}/{destination_blob_name}: {e}")


def move_blob(
    s3: S3Client,
    bucket_name: str,
    blob_name: str,
    destination_bucket_name: str,
    destination_blob_name: str,
):
    """Moves an object by copying to destination and deleting the original."""
    copy_source = {"Bucket": bucket_name, "Key": blob_name}
    try:
        s3.copy_object(Bucket=destination_bucket_name, CopySource=copy_source, Key=destination_blob_name)
        s3.delete_object(Bucket=bucket_name, Key=blob_name)
    except ClientError as e:
        raise RuntimeError(
            f"Failed to move s3://{bucket_name}/{blob_name} to s3://{destination_bucket_name}/{destination_blob_name}: {e}"
        )


def zipdir(path: str, ziph: zipfile.ZipFile):
    # ziph is zipfile handle
    zip_path: Path = Path(path)
    for file in zip_path.glob("*"):
        fi_path: Path = Path(file)
        if fi_path.is_dir() or fi_path.name == ".DS_Store" or fi_path.suffix == ".zip":
            continue
        ziph.write(fi_path, fi_path.name)


def upload_event(
    storm_qgis_filename: Optional[str],
    # bucket: str = "godin-hurricane-data",
    bucket: str = "godin-data",
    s3: Optional[S3Client] = None,
):
    if storm_qgis_filename is None:
        print("Set a qgis filename")
        exit(2)

    if s3 is None:
        s3 = _s3_client(
            aws_access_key_id="GK3d69e8e60c30638ee7bc411d",
            aws_secret_access_key="ffa9d59e420ef2588cdaf0354d8910dc0231e78615466830e677eed1af006a65",
            region_name="garage",
            endpoint_url="http://goose.hive:3900"
        )

    data_path: str = "./data"

    fi_splits: List[str] = storm_qgis_filename.split("_")
    storm_name: str = fi_splits[0]
    filename, _ = storm_qgis_filename.split(".", maxsplit=1)

    # zip relevant files
    zip_output = Path(f"{data_path}/{storm_name}/{filename}.zip")
    zip_output.parent.mkdir(parents=True, exist_ok=True)
    with zipfile.ZipFile(zip_output, "w", zipfile.ZIP_DEFLATED) as zipf:
        zipdir(f"{data_path}/{storm_name}", zipf)

    # Move any existing latest objects to past/
    blobs = list_blobs_with_prefix(
        s3,
        bucket_name=bucket,
        prefix=f"{storm_name}/latest",
    )
    for blob in blobs:
        blob_path: PurePath = PurePath(blob.name)
        if blob.name.endswith("/"):
            continue
        move_blob(
            s3,
            bucket_name=bucket,
            blob_name=blob.name,
            destination_bucket_name=bucket,
            destination_blob_name=f"{storm_name}/past/{blob_path.name}",
        )

    # ZIP data
    source = f"{data_path}/{storm_name}/{filename}.zip"
    zip_dest = f"{storm_name}/latest/{filename}.zip"
    upload_blob(
        s3,
        bucket_name=bucket,
        source_file_name=source,
        destination_blob_name=zip_dest,
    )

    # gis
    source = f"{data_path}/{storm_name}/{filename}.jpeg"
    gis_dest = f"{storm_name}/latest/{filename}.jpeg"
    upload_blob(
        s3,
        bucket_name=bucket,
        source_file_name=source,
        destination_blob_name=gis_dest,
    )

    return zip_dest, gis_dest


if __name__ == "__main__":
    qgis_filename: str = "matthew2016_100x100_2021-09-07T14:37:00+00:00.jpeg"
    upload_event(qgis_filename)
