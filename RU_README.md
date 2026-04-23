# xray-cshare

Языки: [English](README.md) | [Русский](RU_README.md)

`xray-cshare` — это C-совместимая shared library-обертка над [`xray-core`](https://github.com/XTLS/Xray-core). Она предоставляет небольшой ABI для запуска и остановки инстансов Xray из JSON-конфигов, проверки соединения через локальные proxy-порты, получения версии встроенного `xray-core`, а также запуска helper-функций для UUID, TLS, REALITY, VLESS encryption и данных, связанных с сертификатами.

Этот репозиторий ориентирован на интеграторов, которым нужно вызывать Xray из другого runtime через сгенерированную shared library, такую как `.dll`, `.so` или `.dylib`, а не подключать Go-пакеты напрямую.

## Возможности

- Запуск и остановка инстансов `xray-core` из JSON-конфигурации
- Проверка, запущен ли именованный инстанс в данный момент
- Ping целевого URL напрямую или через локальные proxy-порты
- Получение версии встроенного `xray-core`
- Генерация X25519, ML-DSA-65, ML-KEM-768, VLESS encryption helper-значений, UUID и TLS certificate materials
- Вычисление SHA-256 hash цепочки сертификатов с использованием TLS utility из Xray
- Предоставление компактного бинарного формата ответа, удобного для FFI-потребителей

## Сборка

Проект собирается как C-shared library.

Пример локальной сборки:

```bash
go build -buildmode=c-shared -o build/xray_sdk.dll
```

Текущий CI workflow собирает:

- `windows_amd64.dll`
- `linux_amd64.so`
- `linux_arm64.so`
- `darwin_amd64.dylib`
- `darwin_arm64.dylib`
- `darwin_universal.dylib`

Сгенерированный header-файл содержит экспортируемый C ABI. Пример такого header-файла лежит в `build/xray_sdk.h`.

## Краткий обзор

Типичный сценарий использования:

1. Соберите shared library и загрузите ее из вашего host-приложения.
2. Вызовите `Start(uuid, jsonConfig)`, чтобы создать и запустить инстанс Xray, связанный с вашим UUID-ключом.
3. Разберите возвращенный response buffer, чтобы определить успех или ошибку.
4. При необходимости вызовите `IsStarted(uuid)`, `Ping(...)` или `PingConfig(...)`.
5. Вызовите `Stop(uuid)`, когда инстанс больше не нужен.
6. Освободите все возвращенные буферы через `FreePointer`.

## Формат буфера ответа

Большинство экспортируемых функций возвращают raw pointer на буфер, выделенный через `C.malloc`. Формат буфера определен в `transfer/response.go`.

- Байты `0..3`: little-endian `uint32` status code
- Байты `4..5`: little-endian `uint16` content type
- Байты `6..7`: неиспользуемый padding
- Байты начиная с offset `8`: null-terminated UTF-8 строка тела ответа

Значения content type:

- `1`: строка ошибки
- `2`: строка с JSON payload
- `3`: строка обычного сообщения об успехе

Поведение status code:

- `0`: успех
- ненулевое значение: код ошибки

Управление памятью:

- Любой pointer, возвращенный экспортируемой функцией, которая выделяет response buffer, должен освобождаться через `FreePointer`.
- `Stop` и `IsStarted` не выделяют response buffer.

## Коды ошибок

Обертка определяет следующие явные коды в исходниках:

| Code | Name | Meaning |
| --- | --- | --- |
| `1` | `JsonParseError` | Ошибка парсинга JSON-конфига |
| `2` | `LoadConfigError` | Ошибка сборки Xray-конфига |
| `3` | `InitXrayError` | Ошибка `core.New(...)` |
| `4` | `StartXrayError` | Ошибка `instance.Start()` |
| `5` | `XrayAlreadyStarted` | Тот же UUID уже связан с работающим инстансом |
| `6` | `PingTimeoutError` | Объявлен в коде, но сейчас не возвращается публичным API |
| `7` | `PingError` | Объявлен в коде, но сейчас не возвращается публичным API |

Для helper-функций общие ошибки обычно возвращаются как code `1` со строкой ошибки в теле ответа.

## Публичный API

Экспортируемые C-сигнатуры ниже соответствуют `build/xray_sdk.h`.

### Жизненный цикл инстанса

#### `void* Start(char* cUuid, char* cJson);`

Запускает инстанс Xray из JSON-конфига и сохраняет его под UUID-ключом, переданным вызывающей стороной.

Параметры:

- `cUuid`: пользовательский идентификатор, используемый как ключ в map инстансов
- `cJson`: полный Xray JSON-конфиг

Успех:

- Content type `3`
- Тело сообщения: `"Server started"`

Ошибка:

- Возвращает content type `1`
- Код ошибки может быть `1`, `2`, `3`, `4` или `5`

Используемые вызовы `xray-core`:

- `serial.DecodeJSONConfig(strings.NewReader(jsonConfig))`
- `config.Build()`
- `core.New(coreCfg)`
- `instance.Start()`
- `instance.Close()` при cleanup после ошибки запуска

Примечания:

- Если UUID уже существует и связанный инстанс все еще работает, вызов завершится ошибкой с code `5`.
- Успешно созданные инстансы сохраняются в in-memory map, защищенной mutex.

#### `void Stop(char* cUuid);`

Останавливает и удаляет инстанс Xray, сохраненный под указанным UUID.

Параметры:

- `cUuid`: ключ инстанса, использованный в `Start`

Успех:

- Функция ничего не возвращает

Используемые вызовы `xray-core`:

- `instance.Close()`

Примечания:

- Если UUID не найден, функция ничего не делает.
- После остановки Go-обертка также вызывает `runtime.GC()` и `debug.FreeOSMemory()`.

#### `int IsStarted(char* cUuid);`

Проверяет, существует ли сохраненный инстанс и находится ли он в состоянии running.

Параметры:

- `cUuid`: ключ инстанса, использованный в `Start`

Успех:

- Возвращает `1`, если инстанс существует и `instance.IsRunning()` возвращает true
- Возвращает `0` во всех остальных случаях

Используемые вызовы `xray-core`:

- `instance.IsRunning()`

### Проверка соединения

#### `void* PingConfig(char* jsonConfig, int* portsPtr, int portsLen, char* testingURL);`

Запускает временный инстанс Xray из переданного JSON-конфига, затем выполняет HTTP `HEAD` запросы через указанные локальные порты.

Параметры:

- `jsonConfig`: Xray JSON-конфиг для временного инстанса
- `portsPtr`: указатель на массив локальных портов
- `portsLen`: количество портов в `portsPtr`
- `testingURL`: целевой URL для HTTP `HEAD`

Успех:

- Content type `2`
- JSON-массив вида:

```json
[
  {
    "port": 1080,
    "timeout": 123,
    "error": ""
  }
]
```

Ошибка:

- Content type `1`
- Строка с текстом ошибки

Используемые вызовы `xray-core`:

- Косвенно использует тот же startup flow, что и `Start`, через `xray.Start(jsonConfig)`
- Вызывает `instance.Close()` после завершения всех проверок

Примечания:

- Каждый порт проверяется конкурентно.
- Время ответа возвращается в миллисекундах.
- Функция не сохраняет инстанс в глобальной UUID map.

#### `void* Ping(int port, char* testingURL);`

Выполняет прямой HTTP `HEAD` запрос или запрос через proxy `127.0.0.1:<port>`.

Параметры:

- `port`: proxy-порт; `0` означает прямой запрос без proxy
- `testingURL`: целевой URL для HTTP `HEAD`

Успех:

- Content type `2`
- JSON-объект:

```json
{
  "port": 1080,
  "timeout": 123,
  "error": ""
}
```

Поведение при ошибке:

- Функция все равно возвращает payload object
- При ошибке запроса `timeout` обычно равен `-1`, а `error` содержит строку runtime-ошибки

Используемые вызовы `xray-core`:

- Нет прямых вызовов

### Версия

#### `void* GetXrayCoreVersion(void);`

Возвращает строку версии встроенного `xray-core`.

Успех:

- Content type `3`
- Тело сообщения равно результату `core.Version()`

Используемые вызовы `xray-core`:

- `core.Version()`

### Криптографические helper-функции

#### `void* Curve25519Genkey(char* cKey);`

Генерирует или повторно выводит X25519 key material, используя URL-safe base64 без padding.

Параметры:

- `cKey`: необязательный base64-кодированный 32-байтный private key в формате `base64.RawURLEncoding`

Успех:

- Content type `2`
- JSON-объект:

```json
{
  "private_key": "...",
  "password": "...",
  "hash32": "..."
}
```

Ошибка:

- Content type `1`
- Строка ошибки, например при неверной длине private key

Используемые вызовы `xray-core`:

- Нет прямых вызовов

#### `void* Curve25519GenkeyWG(char* cKey);`

Имеет то же поведение, что и `Curve25519Genkey`, но использует стандартный padded base64, подходящий для WireGuard-style представления.

Используемые вызовы `xray-core`:

- Нет прямых вызовов

#### `void* ExecuteUUID(char* cInput);`

Генерирует или нормализует UUID-строку с использованием UUID-пакета Xray.

Параметры:

- `cInput`: пустая строка для случайного UUID или короткий input, который передается в `uuid.ParseString`

Успех:

- Content type `3`
- Тело сообщения содержит UUID-строку

Ошибка:

- Content type `1`
- Строка ошибки, например если длина input превышает 30 байт

Используемые вызовы `xray-core`:

- `uuid.New()`
- `uuid.ParseString(input)`

Примечания:

- Комментарий в исходнике упоминает UUIDv5/VLESS, но фактическая реализация либо генерирует случайный UUID при пустом input, либо парсит переданную строку.

#### `void* ExecuteMLDSA65(char* cInput);`

Генерирует ML-DSA-65 verification material из переданного или случайного seed.

Параметры:

- `cInput`: необязательный 32-байтный seed, закодированный через `base64.RawURLEncoding`

Успех:

- Content type `2`
- JSON-объект:

```json
{
  "seed": "...",
  "verify": "..."
}
```

Используемые вызовы `xray-core`:

- Нет прямых вызовов

#### `void* ExecuteMLKEM768(char* cInput);`

Генерирует ML-KEM-768 encapsulation material из переданного или случайного seed.

Параметры:

- `cInput`: необязательный 64-байтный seed, закодированный через `base64.RawURLEncoding`

Успех:

- Content type `2`
- JSON-объект:

```json
{
  "seed": "...",
  "client": "...",
  "hash32": "..."
}
```

Используемые вызовы `xray-core`:

- Нет прямых вызовов

#### `void* ExecuteVLESSEnc(void);`

Генерирует helper-строки для VLESS encryption и post-quantum вариантов.

Успех:

- Content type `2`
- JSON-объект:

```json
{
  "decryption": "...",
  "encryption": "...",
  "decryption_pq": "...",
  "encryption_pq": "..."
}
```

Используемые вызовы `xray-core`:

- Нет прямых вызовов

### TLS и сертификаты

#### `void* GenerateCert(char* cDomains, char* cCommonName, char* cOrg, int cIsCA, char* cExpire);`

Генерирует PEM-сертификат и пару private key.

Параметры:

- `cDomains`: DNS-имена, разделенные запятыми
- `cCommonName`: common name сертификата
- `cOrg`: имя организации
- `cIsCA`: `1` для генерации CA-style сертификата, иначе `0`
- `cExpire`: необязательная строка длительности в формате Go, например `2160h`

Успех:

- Content type `2`
- JSON-объект:

```json
{
  "certificate": [
    "-----BEGIN CERTIFICATE-----",
    "..."
  ],
  "key": [
    "-----BEGIN PRIVATE KEY-----",
    "..."
  ]
}
```

Ошибка:

- Content type `1`
- Строка ошибки

Используемые вызовы `xray-core`:

- `cert.Generate(nil, opts...)`
- `cert.Authority(true)`
- `cert.KeyUsage(...)`
- `cert.NotAfter(...)`
- `cert.CommonName(...)`
- `cert.DNSNames(...)`
- `cert.Organization(...)`

Примечания:

- Реализация использует пакет генерации TLS-сертификатов из Xray.
- В исходниках предполагается default expiry в 90 дней и default names `"Xray Inc"`, но текущую реализацию следует использовать осторожно, поскольку `initDefaults()` разыменовывает `Expire`, когда он равен `nil`. Этот README описывает фактический путь выполнения, а не предполагает исправленное поведение.

#### `void* ExecuteCertChainHash(char* cCert);`

Вычисляет SHA-256 hash PEM certificate chain с использованием TLS helper из Xray.

Параметры:

- `cCert`: либо PEM-содержимое, либо путь к PEM-файлу в файловой системе

Успех:

- Content type `3`
- Тело сообщения содержит строку hash

Ошибка:

- Content type `1`
- Строка ошибки, если чтение файла завершилось неудачей

Используемые вызовы `xray-core`:

- `tls.CalculatePEMCertChainSHA256Hash(certContent)`

### Служебные функции

#### `void* SetEnv(char* cKey, char* cValue);`

Устанавливает process environment variable внутри host-процесса.

Параметры:

- `cKey`: имя environment variable
- `cValue`: значение environment variable

Успех:

- Content type `3`
- Тело сообщения: `"done"`

Ошибка:

- Content type `1`
- Строка ошибки от `os.Setenv`

Используемые вызовы `xray-core`:

- Нет

#### `void FreePointer(void* ptr);`

Освобождает буфер, ранее возвращенный этой библиотекой.

Параметры:

- `ptr`: pointer, возвращенный одной из функций, создающих response

Успех:

- Функция ничего не возвращает

Примечания:

- Освобождайте только те pointers, которые были выделены этой библиотекой.

## Как эта обертка использует `xray-core`

### Прямая runtime-интеграция

Основной runtime path в `xray/server.go` использует `xray-core` напрямую:

- `infra/conf/serial.DecodeJSONConfig` парсит JSON input
- `config.Build()` преобразует распарсенный конфиг в core config
- `core.New(...)` создает новый Xray instance
- `(*core.Instance).Start()` запускает его
- `(*core.Instance).Close()` останавливает его или выполняет cleanup
- `core.Version()` возвращает версию встроенного core

### Регистрация через blank imports

`xray/imports.go` импортирует множество пакетов `xray-core` ради side effects их `init()`-функций, чтобы handlers, features, proxies, transports и transport headers были зарегистрированы до сборки конфигов.

Сюда входят:

- обязательные app features, такие как dispatcher и proxyman
- optional services, такие как commander, stats, routing, logging, DNS, metrics и observatory
- inbound и outbound proxies, такие как VLESS, VMess, SOCKS, Trojan, Shadowsocks, Freedom, DNS и WireGuard
- transports, такие как gRPC, TCP, KCP, TLS, WebSocket, SplitHTTP, HTTP upgrade, UDP и REALITY
- transport headers, такие как HTTP, SRTP, UTP, WeChat, TLS и WireGuard

Без этих imports загрузка конфига может завершиться ошибкой, потому что компоненты, на которые ссылается JSON, не будут зарегистрированы.

### Использование helper-функций через пакеты `xray-core`

Некоторые helper-функции не относятся к runtime lifecycle, но все равно зависят от пакетов `xray-core`:

- `ExecuteUUID` использует `github.com/xtls/xray-core/common/uuid`
- `GenerateCert` использует `github.com/xtls/xray-core/common/protocol/tls/cert`
- `ExecuteCertChainHash` использует `github.com/xtls/xray-core/transport/internet/tls`
- `crypto_helpers/cert.go` также импортирует `github.com/xtls/xray-core/main/commands/base` для `stringList.Set`

## Пример FFI workflow

Точный host-side код зависит от вашего языка, но ожидаемый lifecycle выглядит так:

```c
void* resp = Start("instance-1", "{...json config...}");
/* read status code, content type, and body from the packed buffer */
FreePointer(resp);

if (IsStarted("instance-1")) {
    void* version = GetXrayCoreVersion();
    /* read packed response */
    FreePointer(version);
}

Stop("instance-1");
```

При декодировании возвращенного pointer:

1. Прочитайте 32-битный little-endian status code из offset `0`.
2. Прочитайте 16-битный little-endian content type из offset `4`.
3. Прочитайте null-terminated body string, начиная с offset `8`.
4. Вызовите `FreePointer`, когда завершите работу с буфером.

## Известные особенности

- Владение инстансами полностью определяется UUID-строкой, которую передает вызывающая сторона.
- Map инстансов существует только в памяти процесса и не сохраняется.
- `PingConfig` создает временный инстанс Xray из переданного конфига, выполняет проверки и затем закрывает его.
- `Ping` возвращает payload object даже при ошибке запроса; в этом случае ошибка встраивается в JSON payload.
- Константы `PingTimeoutError` и `PingError` существуют в коде, но сейчас не отдаются как отдельные response codes через экспортируемый API.
- `GenerateCert` следует использовать осторожно, если `cExpire` пустой, потому что текущий путь реализации вокруг default expiry initialization небезопасен в текущем виде.
- `ExecuteCertChainHash` интерпретирует input как путь к файлу, если такой путь существует на диске; иначе input трактуется как PEM-содержимое.

## Источник истины

Этот README основан на текущей реализации в:

- `main.go`
- `xray/server.go`
- `xray/imports.go`
- `crypto_helpers/*.go`
- `testing/testing.go`
- `transfer/response.go`
- `build/xray_sdk.h`

Если реализация и документация когда-либо разойдутся, считайте исходный код и сгенерированный header authoritative источником.
