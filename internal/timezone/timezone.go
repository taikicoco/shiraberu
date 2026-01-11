package timezone

import "time"

// JST は日本標準時 (UTC+9) のタイムゾーン
var JST = time.FixedZone("JST", 9*60*60)
