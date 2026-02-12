package server

// Route registration is handled in app.go's initServer() method.
// Each domain's RegisterRoutes() function is called with the appropriate
// Echo instance and auth middleware.
//
// Domain route registration pattern:
//   healthhttp.RegisterRoutes(e, healthHandler)
//   userhttp.RegisterRoutes(e, userHandler, authMiddleware)
//   authhttp.RegisterRoutes(e, authHandler, oauthHandler, authMiddleware)
//   wishlisthttp.RegisterRoutes(e, wishlistHandler, authMiddleware)
//   itemhttp.RegisterRoutes(e, itemHandler, authMiddleware)
//   wishlistitemhttp.RegisterRoutes(e, wishlistItemHandler, authMiddleware)
//   reservationhttp.RegisterRoutes(e, reservationHandler, authMiddleware)
//   storagehttp.RegisterRoutes(e, storageHandler, tokenManager)
