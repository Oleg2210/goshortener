File: main
Build ID: cda9eb7b0cc3089a5d4fafa79ebaf1f9b6d1272b
Type: inuse_space
Time: 2026-02-24 04:36:13 +05
Showing nodes accounting for 1.70MB, 12.29% of 13.83MB total
      flat  flat%   sum%        cum   cum%
   -1.50MB 10.85% 10.85%    -1.50MB 10.85%  runtime.malg
    0.92MB  6.68%  4.17%     0.92MB  6.68%  github.com/Oleg2210/goshortener/internal/repository.(*MemoryRepository).Save
    0.53MB  3.85%  0.33%     0.53MB  3.85%  github.com/Oleg2210/goshortener/internal/repository.NewMemoryRepository (inline)
    0.50MB  3.63%  3.30%     0.50MB  3.63%  bufio.NewReaderSize (inline)
    0.50MB  3.63%  6.93%     0.50MB  3.63%  bufio.NewWriterSize (inline)
   -0.50MB  3.62%  3.31%    -0.50MB  3.62%  runtime.allocm
    0.50MB  3.62%  6.93%     0.50MB  3.62%  runtime.acquireSudog
    0.50MB  3.62% 10.54%     0.50MB  3.62%  sync.runtime_notifyListWait
    0.17MB  1.22% 11.76%     0.24MB  1.75%  encoding/json.MarshalIndent
    0.07MB  0.53% 12.29%     0.07MB  0.53%  encoding/json.Marshal
         0     0% 12.29%     0.50MB  3.63%  bufio.NewReader (inline)
         0     0% 12.29%    -1.84MB 13.27%  github.com/Oleg2210/goshortener/internal/handler.(*App).HandlePost
         0     0% 12.29%    -1.84MB 13.27%  github.com/Oleg2210/goshortener/internal/repository.(*FileRepository).Save
         0     0% 12.29%        3MB 21.70%  github.com/Oleg2210/goshortener/internal/repository.(*FileRepository).loadDataFromFile
         0     0% 12.29%     0.24MB  1.75%  github.com/Oleg2210/goshortener/internal/repository.(*FileRepository).saveToFile
         0     0% 12.29%     3.53MB 25.55%  github.com/Oleg2210/goshortener/internal/repository.NewFileRepository
         0     0% 12.29%    -1.84MB 13.27%  github.com/Oleg2210/goshortener/internal/service.(*ShortenerService).Shorten
         0     0% 12.29%    -1.84MB 13.27%  github.com/Oleg2210/goshortener/pkg/middleware/compress.GzipMiddleware.func1
         0     0% 12.29%    -1.84MB 13.27%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 12.29%    -1.84MB 13.27%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0% 12.29%     3.53MB 25.55%  main.chooseStorage
         0     0% 12.29%     3.53MB 25.55%  main.main
         0     0% 12.29%    -1.84MB 13.27%  main.main.AuthMiddleware.func3.1
         0     0% 12.29%    -1.84MB 13.27%  main.main.LoggingMiddleware.func2.1
         0     0% 12.29%    -0.33MB  2.39%  net/http.(*conn).serve
         0     0% 12.29%     0.50MB  3.62%  net/http.(*connReader).abortPendingRead
         0     0% 12.29%     0.50MB  3.62%  net/http.(*response).finishRequest
         0     0% 12.29%    -1.84MB 13.27%  net/http.HandlerFunc.ServeHTTP
         0     0% 12.29%     0.50MB  3.63%  net/http.newBufioReader
         0     0% 12.29%     0.50MB  3.63%  net/http.newBufioWriterSize
         0     0% 12.29%    -1.84MB 13.27%  net/http.serverHandler.ServeHTTP
         0     0% 12.29%     0.50MB  3.62%  runtime.gcBgMarkWorker
         0     0% 12.29%     0.50MB  3.62%  runtime.gcMarkDone
         0     0% 12.29%     3.53MB 25.55%  runtime.main
         0     0% 12.29%    -0.50MB  3.62%  runtime.mstart
         0     0% 12.29%    -0.50MB  3.62%  runtime.mstart0
         0     0% 12.29%    -0.50MB  3.62%  runtime.mstart1
         0     0% 12.29%    -0.50MB  3.62%  runtime.newm
         0     0% 12.29%    -1.50MB 10.85%  runtime.newproc.func1
         0     0% 12.29%    -1.50MB 10.85%  runtime.newproc1
         0     0% 12.29%    -0.50MB  3.62%  runtime.resetspinning
         0     0% 12.29%    -0.50MB  3.62%  runtime.schedule
         0     0% 12.29%     0.50MB  3.62%  runtime.semacquire (inline)
         0     0% 12.29%     0.50MB  3.62%  runtime.semacquire1
         0     0% 12.29%    -0.50MB  3.62%  runtime.startm
         0     0% 12.29%    -1.50MB 10.85%  runtime.systemstack
         0     0% 12.29%    -0.50MB  3.62%  runtime.wakep
         0     0% 12.29%     0.50MB  3.62%  sync.(*Cond).Wait