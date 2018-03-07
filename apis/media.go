package apis

import (
	"fmt"
	"gofreta/app"
	"gofreta/daos"
	"gofreta/models"
	"gofreta/utils"
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/globalsign/mgo"
	routing "github.com/go-ozzo/ozzo-routing"
)

// MediaApi defines media api services
type MediaApi struct {
	routeGroup   *routing.RouteGroup
	mongoSession *mgo.Session
	dao          *daos.MediaDAO
}

// InitMediaApi sets up the routing of media endpoints and the corresponding handlers.
func InitMediaApi(rg *routing.RouteGroup, session *mgo.Session) {
	api := MediaApi{
		routeGroup:   rg,
		mongoSession: session,
		dao:          daos.NewMediaDAO(session),
	}

	rg.Get("/media", authenticateToken(session, "media", "index"), api.index)
	rg.Post("/media", authenticateToken(session, "media", "upload"), api.upload)
	rg.Get("/media/<id>", authenticateToken(session, "media", "view"), api.view)
	rg.Put("/media/<id>", authenticateToken(session, "media", "update"), api.update)
	rg.Post("/media/<id>", authenticateToken(session, "media", "update"), api.replace)
	rg.Delete("/media/<id>", authenticateToken(session, "media", "delete"), api.delete)
}

// index api handler for fetching paginated media items list
func (api *MediaApi) index(c *routing.Context) error {
	// --- fetch search data
	searchFields := []string{"title", "type", "path", "created", "modified"}
	searchData := utils.GetSearchConditions(c, searchFields)
	// ---

	// --- fetch sort data
	sortFields := []string{"title", "type", "path", "created", "modified"}
	sortData := utils.GetSortFields(c, sortFields)
	// ---

	total, _ := api.dao.Count(searchData)

	limit, page := utils.GetPaginationSettings(c, total)

	utils.SetPaginationHeaders(c, limit, total, page)

	items := []models.Media{}

	if total > 0 {
		items, _ = api.dao.GetList(limit, limit*(page-1), searchData, sortData)

		items = daos.ToAbsMediaPaths(items)
	}

	return c.Write(items)
}

// view api handler for fetching single media item data
func (api *MediaApi) view(c *routing.Context) error {
	id := c.Param("id")

	model, err := api.dao.GetByID(id)

	if err != nil {
		return utils.NewNotFoundError(fmt.Sprintf("Media item with id \"%v\" doesn't exist!", id))
	}

	model = daos.ToAbsMediaPath(model)

	return c.Write(model)
}

// update api handler for updating existing media item settings (eg. name)
func (api *MediaApi) update(c *routing.Context) error {
	id := c.Param("id")

	model, fetchErr := api.dao.GetByID(id)
	if fetchErr != nil {
		return utils.NewNotFoundError(fmt.Sprintf("Media item with id \"%v\" doesn't exist!", id))
	}

	form := &models.MediaUpdateForm{}
	if readErr := c.Read(form); readErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while updating media item.", readErr)
	}

	form.Model = model

	updatedModel, updateErr := api.dao.Update(form)

	if updateErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while updating media item.", updateErr)
	}

	updatedModel = daos.ToAbsMediaPath(updatedModel)

	return c.Write(updatedModel)
}

// delete api handler for deleting an existing media item
func (api *MediaApi) delete(c *routing.Context) error {
	id := c.Param("id")

	model, fetchErr := api.dao.GetByID(id)
	if fetchErr != nil {
		return utils.NewNotFoundError(fmt.Sprintf("Media item with id \"%v\" doesn't exist!", id))
	}

	deleteErr := api.dao.Delete(model)
	if deleteErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while deleting Language item.", deleteErr)
	}

	c.Response.WriteHeader(http.StatusNoContent)

	return nil
}

// upload api handler for uploading and creating multiple media items
func (api *MediaApi) upload(c *routing.Context) error {
	items := []*models.Media{}
	errs := map[string]error{}

	//get the multipart reader for the request.
	reader, readerErr := c.Request.MultipartReader()
	if readerErr != nil {
		return readerErr
	}

	// copy each part to destination
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}

		// if no file name is provided, skip the iteration
		if part.FileName() == "" {
			continue
		}

		name := part.FileName()
		ext := filepath.Ext(name)
		basename := strings.TrimSuffix(name, ext)

		uploadPath, uploadFileType, uploadErr := uploadFile(part)
		if uploadErr != nil {
			errs[name] = uploadErr
			continue
		}

		mediaData := &models.Media{
			Title: basename,
			Type:  uploadFileType,
			// store only the name of the file for easier migrations
			Path: filepath.Base(uploadPath),
		}

		// db record create
		model, createErr := api.dao.Create(mediaData)
		if createErr != nil {
			errs[name] = createErr
			continue
		}

		model = daos.ToAbsMediaPath(model)

		items = append(items, model)
	}

	data := &struct {
		Items  []*models.Media  `json:"items"`
		Errors map[string]error `json:"errors"`
	}{items, errs}

	return c.Write(data)
}

// replace api handler for replacing existing media file
func (api *MediaApi) replace(c *routing.Context) error {
	id := c.Param("id")

	model, fetchErr := api.dao.GetByID(id)
	if fetchErr != nil {
		return utils.NewNotFoundError(fmt.Sprintf("Media item with id \"%v\" doesn't exist!", id))
	}
	oldModel := *model

	file, _, readErr := c.Request.FormFile("file")
	if readErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while replacing media file.", readErr)
	}
	defer file.Close()

	filePath, fileType, uploadErr := uploadFile(file)
	if uploadErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while replacing media file.", uploadErr)
	}

	model.Type = fileType
	model.Path = filepath.Base(filePath) // store only the name of the file for easier migrations

	// db record update
	updatedModel, updateErr := api.dao.Replace(model)
	if updateErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while replacing media file.", updateErr)
	}

	// remove old file
	oldModel.DeleteFile()

	updatedModel = daos.ToAbsMediaPath(updatedModel)

	return c.Write(updatedModel)
}

// uploadFile validates and uploads single file from a reader (returns the uploaded file path and type on success).
func uploadFile(r io.Reader) (string, string, error) {
	b, readErr := ioutil.ReadAll(r)
	if readErr != nil {
		return "", "", readErr
	}

	// --- validate
	isValid := utils.ValidateMimeType(b, models.ValidMediaTypes())
	if isValid != true {
		return "", "", utils.NewDataError("Invalid or unsupported file type.")
	}

	sizeErr := utils.ValidateSize(b, app.Config.GetFloat64("upload.maxSize"))
	if sizeErr != true {
		return "", "", utils.NewDataError("Media is too big.")
	}
	// ---

	// --- normalize
	ext, fileType := utils.GetExtAndFileTypeByMimeType(b)
	name := utils.MD5(utils.Random(10)+time.Now().String()) + "." + ext

	uploadDir := strings.TrimSuffix(app.Config.GetString("upload.dir"), "/")
	uploadPath := uploadDir + "/" + name

	// ensure that the upload directory exist
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, 0777)
	}
	// ---

	// save file
	dst, dstCreateErr := os.Create(uploadPath)
	if dstCreateErr != nil {
		return "", "", dstCreateErr
	}
	defer dst.Close()

	_, writeErr := dst.Write(b)
	if writeErr != nil {
		return "", "", writeErr
	}

	if fileType == utils.FILE_TYPE_IMAGE {
		go createThumbs(uploadPath)
	}

	return uploadPath, fileType, nil
}

// createThumbs creates image file thumbs.
func createThumbs(file string) error {
	img, err := imaging.Open(file)
	if err != nil {
		return err
	}

	ext := filepath.Ext(file)
	basePath := strings.TrimSuffix(file, ext)

	thumbSizes := app.Config.GetStringSlice("upload.thumbs")
	for _, size := range thumbSizes {
		parts := strings.SplitN(size, "x", 2)
		if len(parts) != 2 {
			continue
		}

		w, _ := strconv.Atoi(parts[0])
		h, _ := strconv.Atoi(parts[1])

		thumb := imaging.Thumbnail(img, w, h, imaging.CatmullRom)

		// create a new blank image
		dst := imaging.New(w, h, color.NRGBA{0, 0, 0, 0})

		// paste thumbnail into the new image
		dst = imaging.Paste(dst, thumb, image.Pt(0, 0))

		// save to file
		if err := imaging.Save(dst, basePath+"_"+size+ext); err != nil {
			return err
		}
	}

	return nil
}
