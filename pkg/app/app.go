package app

const MODULE = "app"

/* Environment variables requried by the Lambda function:
GEOIP_DATABSE_PATH: path to the GeoIP database file provided by a Lambda layer.
MASTODON_ACCESS_TOKEN: access token for the Mastodon account.
MASTODON_INSTANCE_URL: URL of the Mastodon instance. (e.g. https://mastodon.social)
MASTODON_SUSPEND_TEXT: text to include in the suspension notice.
PSK: pre-shared key, you know... for security.
*/
