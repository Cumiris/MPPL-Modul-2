package http

import (
	"MPPL-Modul-2/manage_product"
	. "MPPL-Modul-2/manage_product/delivery/utils"
	. "MPPL-Modul-2/models/manage_product"
	"github.com/gorilla/mux"
	"github.com/labstack/echo"
	"gopkg.in/go-playground/validator.v9"
	"io/ioutil"
	"net/http"
	"strconv"
)

type ImageHandler struct {
	ImageUsecase manage_product.UseCase
}

func NewImageHandler(e *echo.Echo, imageusecase manage_product.UseCase) {
	handler := &ImageHandler{ImageUsecase: imageusecase}

	e.GET("/images", handler.FetchImage)
	e.GET("/images/:id", handler.GetById)
	e.POST("/images", handler.Store)
	e.PUT("/images/:id", handler.Update)
	e.DELETE("/images/:id", handler.Delete)

	router := mux.NewRouter()

	router.HandleFunc("/images/upload-image", handler.uploadImage())

	//log.Fatal(http.ListenAndServe(":8090", router))
}


func (ph *ImageHandler) FetchImage(c echo.Context) error {
	listEl, err := ph.ImageUsecase.FetchImage()

	if err != nil {
		return c.JSON(GetStatusCode(err), ResponseError{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, listEl)
}

func (ph *ImageHandler) GetById(c echo.Context) error {

	id_, err := strconv.Atoi(c.Param("id"))
	id := uint(id_)

	el, err := ph.ImageUsecase.GetByIdImage(id)

	if err != nil {
		return c.JSON(GetStatusCode(err), ResponseError{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, el)
}

func (ph *ImageHandler) Update(c echo.Context) error {
	var image_ Image
	id_, err := strconv.Atoi(c.Param("id"))
	id := uint(id_)

	image_.ID = id
	err = c.Bind(&image_)

	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}

	if ok, err := isRequestForImageValid(&image_); !ok {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	err = ph.ImageUsecase.UpdateImage(&image_,id)

	if err != nil {
		return c.JSON(GetStatusCode(err), ResponseError{Message: err.Error()})
	}

	return c.JSON(http.StatusCreated, image_)
}

func (ph *ImageHandler) Store(c echo.Context) error {
	var image_ Image

	err := c.Bind(&image_)

	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}

	if ok, err := isRequestForImageValid(&image_); !ok {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	err = ph.ImageUsecase.StoreImage(&image_)

	if err != nil {
		return c.JSON(GetStatusCode(err), ResponseError{Message: err.Error()})
	}

	return c.JSON(http.StatusCreated, image_)
}

func (ph *ImageHandler) Delete(c echo.Context) error {
	id_, err := strconv.Atoi(c.Param("id"))
	id := uint(id_)

	err = ph.ImageUsecase.DeleteImage(id)

	if err != nil {
		return c.JSON(GetStatusCode(err), ResponseError{Message: err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}

func isRequestForImageValid(p *Image) (bool, error) {
	validate := validator.New()

	err := validate.Struct(p)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (ph *ImageHandler) uploadImage() http.HandlerFunc {
	const maxUploadSize = 2 * 1024 * 1024 // 2 mb

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//validate file size


		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			renderError(w, "FILE_TOO_BIG", http.StatusBadRequest)
			return
		}

		// parse and validate file and post parameters
		//fileType := r.PostFormValue("type")
		file, _, err := r.FormFile("uploadFile")
		if err != nil {
			renderError(w, "INVALID_FILE", http.StatusBadRequest)
			return
		}
		defer file.Close()
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			renderError(w, "INVALID_FILE", http.StatusBadRequest)
			return
		}

		// check file type, detectcontenttype only needs the first 512 bytes
		filetype := http.DetectContentType(fileBytes)

		switch filetype {
		case "image/jpeg", "image/jpg":
		case "image/gif", "image/png":
		case "application/pdf":
			break
		default:
			renderError(w, "INVALID_FILE_TYPE", http.StatusBadRequest)
			return
		}

		filename, err := ph.ImageUsecase.UploadImage(fileBytes, filetype)

		if err != nil {
			renderError(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.Write([]byte(filename))
		}
	})


}

func renderError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(message))
}