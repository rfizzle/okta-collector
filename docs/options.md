# Configuring The Collector Binary

For information on how to run the okta-collector binary, detailed usage information can be found by running okta-collector --help. This document is a more detailed version of the information presented in the help output text.

## Options

### How do you specify options?

In order of precedence, options can be specified via:
* Flag
* Environment Variable
* Config

For example, all the following ways of launching Okta-Collector are equivalent:

*Using only CLI flags*

```
$ /usr/bin/okta-collector \
  --okta-domain acme.okta.com \
  --okta-api-key ABC123 \
  --state-path /etc/okta-collector/collector.state \
  --file \
  --file-rotate \
  --file-path /var/log/okta-collector.log
```

*Using only environment variables*

```
$ OC_OKTA_DOMAIN=acme.okta.com \
  OC_OKTA_API_KEY=ABC123 \
  OC_STATE_PATH=/etc/okta-collector/collector.state \
  OC_FILE=true \
  OC_FILE_ROTATE=true \
  OC_FILE_PATH=/var/log/okta-collector.log \
  /usr/bin/okta-collector
```

*Using a config file*

```
$ echo '
{
  "okta-domain": "acme.okta.com",
  "okta-api-key": "ABC123",
  "state-path": "/etc/okta-collector/collector.state",
  "file": true,
  "file-rotate": true,
  "file-path": "/var/log/okta-collector.log"
}
' > /etc/okta-collector/config.json
$ /usr/bin/okta-collector -c --config-file /etc/okta-collector/config.json
```

### What are the options?

Note that all option names can be converted consistently from flag name to environment variable to config file and
visa-versa. For example, the `--okta-domain` flag would be the `OC_OKTA_DOMAIN`. Further, specifying the
`okta-domain` option in the config would follow the pattern:

```
{
  "okta-domain": "acme.okta.com"
...
```

#### General Options

##### `okta-domain` **required**

The organization domain for Okta.

* Default Value: none
* Type: String
* Environment Variable: `OC_OKTA_DOMAIN`
* Config file format (depends on type, presented is JSON):
```
 "okta-domain": "acme.okta.com"
``` 

#### `okta-api-key` **required**

The API key generated to authenticate the collector.

* Default Value: none
* Type: String
* Environment Variable: `OC_OKTA_API_KEY`
* Config file format (depends on type, presented is JSON):
```
 "okta-api-key": "ABC123"
```

#### `schedule`

Time in seconds to run collection job.

* Default Value: 30
* Type: Integer
* Environment Variable: `OC_SCHEDULE`
* Config file format (depends on type, presented is JSON):
```
 "schedule": 60
```

#### `state-path` **required**

The path to the state file where the last poll timestamp will be stored.

* Default Value: `collector.state`
* Type: String
* Environment Variable: `OC_STATE_PATH`
* Config file format (depends on type, presented is JSON):
```
 "state-path": "/etc/okta-collector/collector.state"
```

#### Output Options

#### `file`

This flag will enable writing the logs to a file.

* Default Value: `false`
* Type: Boolean
* Environment Variable: `OC_FILE`
* Config file format (depends on type, presented is JSON):
```
 "file": false
```

#### `file-rotate`

This flag will enable rotation of the written log file after every collection. Useful for ingesting into systems such
as logstash.

* Default Value: `false`
* Type: Boolean
* Environment Variable: `OC_FILE_ROTATE`
* Config file format (depends on type, presented is JSON):
```
 "file-rotate": false
```

#### `file-path`

The destination path to write the log file. If `file-rotate` is enabled, the format will be `{path}.{timestamp}`

* Default Value: none
* Type: String
* Environment Variable: `OC_FILE_PATH`
* Config file format (depends on type, presented is JSON):
```
 "file-path": "/var/log/okta-collector.log"
```

#### `gcs`

This flag will enable writing the logs to Google Cloud Storage.

* Default Value: `false`
* Type: Boolean
* Environment Variable: `OC_GCS`
* Config file format (depends on type, presented is JSON):
```
 "gcs": false
```

#### `gcs-bucket`

The Google Cloud Storage bucket to write to.

* Default Value: none
* Type: String
* Environment Variable: `OC_GCS_BUCKET`
* Config file format (depends on type, presented is JSON):
```
 "gcs-bucket": "acme-logs"
```

#### `gcs-path`

The destination path to write to in Google Cloud Storage.

* Default Value: none
* Type: String
* Environment Variable: `OC_GCS_PATH`
* Config file format (depends on type, presented is JSON):
```
 "gcs-path": "logs/okta/collector.log"
```

##### `gcs-credentials`

The Google JSON key for an IAM account that has been granted write access to the specified bucket.

* Default Value: none
* Type: String
* Environment Variable: `OC_GCS_CREDENTIALS`
* Config file format (depends on type, presented is JSON):
```
 "gcs-credentials": "path/to/credentials/file"
``` 

#### `stackdriver`

This flag will enable writing the logs to Stackdriver.

* Default Value: `false`
* Type: Boolean
* Environment Variable: `OC_STACKDRIVER`
* Config file format (depends on type, presented is JSON):
```
 "stackdriver": false
```

#### `stackdriver-project`

This GCP project that the Stackdriver instance exists in.

* Default Value: none
* Type: String
* Environment Variable: `OC_STACKDRIVER_PROJECT`
* Config file format (depends on type, presented is JSON):
```
 "stackdriver-project": "acme-logging"
```

#### `stackdriver-log-name`

The resource name of the log. This is useful for separating log types in Stackdriver.

* Default Value: none
* Type: String
* Environment Variable: `OC_STACKDRIVER_LOG_NAME`
* Config file format (depends on type, presented is JSON):
```
 "stackdriver-log-name": "okta-system-logs"
```

##### `stackdriver-credentials`

The Google JSON key for an IAM account that has been granted write access to stackdriver.

* Default Value: none
* Type: String
* Environment Variable: `OC_STACKDRIVER_CREDENTIALS`
* Config file format (depends on type, presented is JSON):
```
 "stackdriver-credentials": "path/to/credentials/file"
```

#### `s3`

This flag will enable writing the logs to AWS S3.

* Default Value: `false`
* Type: Boolean
* Environment Variable: `OC_S3`
* Config file format (depends on type, presented is JSON):
```
 "s3": false
```

#### `s3-region`

The region of the AWS S3 bucket.

* Default Value: none
* Type: String
* Environment Variable: `OC_S3_REGION`
* Config file format (depends on type, presented is JSON):
```
 "s3-region": "us-east-2"
```

#### `s3-bucket`

The AWS S3 bucket to write to.

* Default Value: none
* Type: String
* Environment Variable: `OC_S3_BUCKET`
* Config file format (depends on type, presented is JSON):
```
 "s3-bucket": "acme-logs"
```

#### `s3-path`

The destination path to write to in AWS S3.

* Default Value: none
* Type: String
* Environment Variable: `OC_S3_PATH`
* Config file format (depends on type, presented is JSON):
```
 "s3-path": "logs/okta/collector.log"
```

#### `s3-access-key-id`

The AWS S3 access key ID with permission to write to the targeted S3 bucket.

* Default Value: none
* Type: String
* Environment Variable: `OC_S3_ACCESS_KEY_ID`
* Config file format (depends on type, presented is JSON):
```
 "s3-access-key-id": "A1234567890"
```

#### `s3-secret-key`

The AWS S3 secret key of the AWS S3 access key ID.

* Default Value: none
* Type: String
* Environment Variable: `OC_S3_SECRET_KEY`
* Config file format (depends on type, presented is JSON):
```
 "s3-secret-key": "aBcDeFg123"
```

#### `s3-storage-class`

The AWS S3 storage class to save the log files under. This is useful if you are writing to S3 for backups or long term
storage.

* Default Value: `STANDARD`
* Type: String
* Environment Variable: `OC_S3_STORAGE_CLASS`
* Config file format (depends on type, presented is JSON):
```
 "s3-storage-class": "STANDARD"
```

Supported options: ["STANDARD", "REDUCED_REDUNDANCY", "STANDARD_IA", "ONEZONE_IA", "GLACIER", "DEEP_ARCHIVE"]

Read more about S3 storage classes [here](https://aws.amazon.com/s3/storage-classes).

#### `http`

This flag will enable writing the logs to an HTTP endpoint.

* Default Value: `false`
* Type: Boolean
* Environment Variable: `OC_HTTP`
* Config file format (depends on type, presented is JSON):
```
 "http": false
```

You can see an example of the HTTP request [here](./http-example.md)

#### `http-url` **required if HTTP enabled**

The target HTTP endpoint for log submission.

* Default Value: none
* Type: String
* Environment Variable: `OC_HTTP_URL`
* Config file format (depends on type, presented is JSON):
```
 "http-url": false
```

#### `http-auth`

An optional Authorization token to append to the request headers on log submission.

* Default Value: none
* Type: String
* Environment Variable: `OC_HTTP_AUTH`
* Config file format (depends on type, presented is JSON):
```
 "http-auth": "Bearer ABC123"
```

#### `http-max-items`

The maximum number of logs to submit in a single HTTP request. Useful for limiting or batch processing logs in a single
endpoint.

* Default Value: `100`
* Type: Integer
* Environment Variable: `OC_HTTP_MAX_ITEMS`
* Config file format (depends on type, presented is JSON):
```
 "http-max-items": 500
```