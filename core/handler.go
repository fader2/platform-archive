package core

import (
	"bytes"
	"net/http"

	"github.com/fader2/platform/config"
	"github.com/fader2/platform/consts"

	"log"

	"io"

	"github.com/CloudyKit/jet"
	"github.com/fader2/platform/addons"
	"github.com/julienschmidt/httprouter"
	lua "github.com/yuin/gopher-lua"
)

var (
	DefNotFoundTplHandler = func(
		w http.ResponseWriter,
		r *http.Request,
		ps httprouter.Params,
	) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Not found template"))
	}
)

type AssetsLoader interface {
	Open(name string) (io.ReadCloser, error)
	Exists(name string) (string, bool)
}

func EntrypointHandler(
	assets AssetsLoader,
	cfg *config.Config,
	route config.Route,
	tpls *jet.Set,
) func(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
) {
	return func(
		w http.ResponseWriter,
		r *http.Request,
		ps httprouter.Params,
	) {
		if config.IsMaintenance() {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Header().Set("Retry-After", "120")
			w.Write([]byte("Maintenance. Please retry after 120 sec."))
			return
		}

		/*
			- find template
			- setup context and lua executor
				- setup request params
			- execute lua
			- execute tpl
		*/

		var withoutView = len(route.Handler) == 0
		var view *jet.Template

		// Find tpl
		if !withoutView {
			var errGetTpl error
			view, errGetTpl = tpls.GetTemplate(route.Handler)
			if errGetTpl != nil {
				log.Println("find template", route.Handler, errGetTpl)
				// not found template
				DefNotFoundTplHandler(w, r, ps)
				return
			}
		}

		// setup ctx and lua engine
		L := lua.NewState()
		defer L.Close()

		consts.SetupToLua(L)
		addons.PreloadLuaModules(L)

		vars := make(jet.VarMap)
		ctx := NewContext(
			&route,
			w,
			r,
			vars,
		)
		luaGlobalHelpFuncs(L)
		registerContextType(L, ctx)
		registerUserType(L)
		registerAccessTokenType(L)
		config.RegisterReadOnlyConfigType(L, cfg)

		// set request options
		for _, param := range ps {
			vars.Set(param.Key, param.Value)
		}

		// execute all middlewares
		for _, middleware := range route.Middlewares {
			fullPathToMiddleware, exists := assets.Exists(middleware)
			if !exists {
				log.Printf(
					"error open middleware %s, not exists",
					middleware,
				)
				continue
			}
			f, err := assets.Open(fullPathToMiddleware)
			if err != nil {
				log.Printf(
					"error open middleware %s, %s",
					middleware,
					err,
				)
				continue
			}
			defer f.Close()

			d := new(bytes.Buffer)
			io.Copy(d, f)
			if err := L.DoString(d.String()); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal server"))
				log.Printf(
					"error execute middleware %s, %s",
					middleware,
					err,
				)
				return
			}

			if ctx.AbortFromMiddleware {
				break
			}
		}

		if ctx.Abort {
			return
		}

		if view == nil {
			DefNotFoundTplHandler(w, r, ps)
			return
		}

		if ctx.ResponseStatus == -1 {
			ctx.ResponseStatus = http.StatusOK
		}
		w.WriteHeader(ctx.ResponseStatus)
		if err := view.Execute(w, vars, ctx); err != nil {
			log.Println("execute tpl:", err)
		}

		return
	}
}
