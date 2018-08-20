package main

import (
    "encoding/json"
    "log"
    "os"
    "strconv"
    "time"
    "net/http"
    "github.com/ahmdrz/goinsta"
    "github.com/patrickmn/go-cache"
    "github.com/gin-gonic/gin"
)

// Create a cache with a default expiration time of 15 minutes
var c = cache.New(15*time.Minute, 30*time.Minute)

func main() {}

// This function's name is a must. App Engine uses it to drive the requests properly.
func init() {
    // Starts a new Gin instance with no middle-ware
    r := gin.New()

    // Define your handlers
    r.GET("/", func(c *gin.Context) {
        c.String(http.StatusOK, "Hello World!")
    })
    r.GET("/ping", func(c *gin.Context) {
        c.String(http.StatusOK, "pong")
    })
    r.POST("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "pong",
        })
    })
    r.GET("/user", func(c *gin.Context) {
        user := os.Getenv("USERNAME")
        c.String(http.StatusOK, "Hello %s", user)
    })
    // Parameters in querystring
    // Query string parameters are parsed using the existing underlying request object.
    // The request responds to a url matching:  /welcome?firstname=Jane&lastname=Doe
    r.GET("/instagram", func(c *gin.Context) {
        user := os.Getenv("USERNAME")
        password := os.Getenv("PASSWORD")
        limit := c.DefaultPostForm("limit", "25")
        if len(limit) < 1  {
            limit = "25"
        }
        lmt, _ := strconv.Atoi(limit)
        insta, err := login(user, password)
        if err != nil {
            log.Println(err.Error())
            c.String(http.StatusOK, err.Error())
        }
        instagram := instagram(*insta, lmt)

        // c.JSON(200, instagram) // instagram here should be struct not JSON string
        c.String(http.StatusOK, instagram) // This works but isn't kosher
    })
    r.POST("/instagram", func(c *gin.Context) {
        // post form not querystring
        user := c.PostForm("user")
        if len(user) < 1  {
            user = os.Getenv("USERNAME")
            log.Println("user:", user)
        }
        password := c.PostForm("pwd")
        if len(password) < 1  {
            password = os.Getenv("PASSWORD")
        }
        limit := c.DefaultPostForm("limit", "25")
        if len(limit) < 1  {
            limit = "25"
        }
        lmt, _ := strconv.Atoi(limit)
        insta, err := login(user, password)
        if err != nil {
            log.Println(err.Error())
            c.String(http.StatusOK, err.Error())
        }
        instagram := instagram(*insta, lmt)

        c.String(http.StatusOK, instagram) // This works but isn't kosher
    })
    r.Run() // listen and serve on 0.0.0.0:8080
    // Handle all requests using net/http
    http.Handle("/", r)
}


/* login
    returns goinsta.Instagram object
    based on saved JSON object or via new login for user
    TODO - better edge cases
*/
func login(user string, password string) (*goinsta.Instagram, error) {
    var insta *goinsta.Instagram
    gc, found := c.Get(user)
    if found {
        log.Println("Found session", user)
        insta = gc.(*goinsta.Instagram)
    } else {
        log.Println("Not found session", user, "Logging with user/password")
        insta = goinsta.New(user, password)
        err := insta.Login()
        if err != nil {
            log.Println(err.Error())
            return insta, err
        }
        c.Set(user, insta, cache.DefaultExpiration)
    }

    return insta, nil
}


/* instagram
    returns JSON with images metadata (links, places, likers etc.)
    returns <= limit images
    processing is slow, takes to long for AWS Proxy timeout
*/
func instagram(insta goinsta.Instagram, limit int) string {
    var Images []instaImage
    media := insta.Account.Feed()
    i := 0
// Label break (break out of two loops with single break statement)
MediaLoop: 
    for media.Next() { // 2-step iteration 1) Going through pages with Next()
        for _, item := range media.Items { // 2) Iterating through items in a page
            i++
            if len(item.Images.Versions) > 0 {
                // Cast image metadata into smaller object
                Image := cast(item)
                // tm := time.Unix(Image.TakenAt, 0)
                // log.Println(i, ":", Image.ID, "-", tm)
                // Append image to array
                Images = append(Images, Image)
            }
            if i >= limit { break MediaLoop } // We only need so many images
        }
    }
    // Create JSON object from Images
    jsonImages, jsonErr3 := json.MarshalIndent(Images, "    ", "    ")
    if jsonErr3 != nil {
        log.Println(jsonErr3.Error())
    }

    return string(jsonImages)
}

/* cast - cast struct into JSON, into smaller struct */
func cast(item interface{}) instaImage {
    var Image instaImage
    // create JSON from item
    jsonMedia, jsonErr1 := json.MarshalIndent(item, "    ", "    ")
    if jsonErr1 != nil {
        panic (jsonErr1.Error())
    }
    // Unmarshal JSON into Image
    jsonErr2 := json.Unmarshal(jsonMedia, &Image)
    if jsonErr2 != nil {
        panic (jsonErr2.Error())
    }
    return Image
}

/* instaImage   
    Instagram Image striped down */
type instaImage struct {
    TakenAt         int64  `json:"taken_at"`
    ID              string `json:"id"`
    DeviceTimestamp int64  `json:"device_timestamp"`
    MediaType       int    `json:"media_type"`
    ClientCacheKey  string `json:"client_cache_key"`
    Caption         struct {
        Text string `json:"text"`
        User struct {
            Username string `json:"username"`
        } `json:"user,omitempty"`
    } `json:"caption"`
    LikeCount      int      `json:"like_count"`
    TopLikers      []string `json:"top_likers,omitempty"`
    ImageVersions2 struct {
        Candidates []struct {
            Width  int    `json:"width"`
            Height int    `json:"height"`
            URL    string `json:"url"`
        } `json:"candidates"`
    } `json:"image_versions2"`
    OriginalWidth  int `json:"original_width"`
    OriginalHeight int `json:"original_height"`
    Location       struct {
        Name      string `json:"name"`
        City      string `json:"city"`
        ShortName string `json:"short_name"`
        Lng       float64    `json:"lng"`
        Lat       float64    `json:"lat"`
    } `json:"location,omitempty"`
}
