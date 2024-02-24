package logger

import "go.uber.org/zap"

// New создает новый экземпляр логгера с заданным уровнем.
func New(level string) (*zap.Logger, error) {
	// Преобразуем строку уровня логирования в объект AtomicLevel.
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}
	// Создаем конфигурацию логгера для разработки.
	cfg := zap.NewDevelopmentConfig()
	// Устанавливаем уровень логирования в заданный уровень.
	cfg.Level = lvl
	// Строим экземпляр логгера на основе конфигурации.
	zapLogger, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	return zapLogger, nil
}
