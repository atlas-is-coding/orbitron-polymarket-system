package i18n

// Locale holds all translatable strings for the bot UI and log messages.
type Locale struct {
	// Tabs
	TabOverview    string `json:"tab_overview"`
	TabOrders      string `json:"tab_orders"`
	TabPositions   string `json:"tab_positions"`
	TabCopytrading string `json:"tab_copytrading"`
	TabLogs        string `json:"tab_logs"`
	TabSettings    string `json:"tab_settings"`

	// App header
	AppRunning string `json:"app_running"`
	AppWallet  string `json:"app_wallet"`
	HelpGlobal string `json:"help_global"`

	// Overview
	OverviewSubsystems  string `json:"overview_subsystems"`
	OverviewStats       string `json:"overview_stats"`
	OverviewActive      string `json:"overview_active"`
	OverviewInactive    string `json:"overview_inactive"`
	OverviewBalance     string `json:"overview_balance"`
	OverviewOpenOrders  string `json:"overview_open_orders"`
	OverviewPositions   string `json:"overview_positions"`
	OverviewPnLToday    string `json:"overview_pnl_today"`
	OverviewCopyTraders string `json:"overview_copy_traders"`

	// Orders tab
	OrdersColMarket string `json:"orders_col_market"`
	OrdersColSide   string `json:"orders_col_side"`
	OrdersColPrice  string `json:"orders_col_price"`
	OrdersColSize   string `json:"orders_col_size"`
	OrdersColFilled string `json:"orders_col_filled"`
	OrdersColStatus string `json:"orders_col_status"`
	OrdersColAge    string `json:"orders_col_age"`
	OrdersHelp      string `json:"orders_help"`

	// Positions tab
	PosColMarket  string `json:"pos_col_market"`
	PosColSide    string `json:"pos_col_side"`
	PosColSize    string `json:"pos_col_size"`
	PosColEntry   string `json:"pos_col_entry"`
	PosColCurrent string `json:"pos_col_current"`
	PosColPnL     string `json:"pos_col_pnl"`
	PosColPnLPct  string `json:"pos_col_pnl_pct"`
	PosHelp       string `json:"pos_help"`

	// Copytrading tab
	CopyColAddress   string `json:"copy_col_address"`
	CopyColLabel     string `json:"copy_col_label"`
	CopyColStatus    string `json:"copy_col_status"`
	CopyColAlloc     string `json:"copy_col_alloc"`
	CopyTraders      string `json:"copy_traders"`
	CopyRecentTrades string `json:"copy_recent_trades"`
	CopyNoData       string `json:"copy_no_data"`

	// Logs tab
	LogsFrozen string `json:"logs_frozen"`
	LogsFilter string `json:"logs_filter"`
	LogsHelp   string `json:"logs_help"`

	// Settings sections
	SectionUI            string `json:"section_ui"`
	SectionAuth          string `json:"section_auth"`
	SectionAPI           string `json:"section_api"`
	SectionMonitor       string `json:"section_monitor"`
	SectionTradesMonitor string `json:"section_trades_monitor"`
	SectionTrading       string `json:"section_trading"`
	SectionCopytrading   string `json:"section_copytrading"`
	SectionTelegram      string `json:"section_telegram"`
	SectionDatabase      string `json:"section_database"`
	SectionLog           string `json:"section_log"`

	// Settings field labels
	FieldLanguage       string `json:"field_language"`
	FieldPrivKey        string `json:"field_priv_key"`
	FieldAPIKey         string `json:"field_api_key"`
	FieldAPISecret      string `json:"field_api_secret"`
	FieldPassphrase     string `json:"field_passphrase"`
	FieldChainID        string `json:"field_chain_id"`
	FieldTimeout        string `json:"field_timeout"`
	FieldMaxRetries     string `json:"field_max_retries"`
	FieldEnabled        string `json:"field_enabled"`
	FieldPollInterval   string `json:"field_poll_interval"`
	FieldAlertOnFill    string `json:"field_alert_on_fill"`
	FieldAlertOnCancel  string `json:"field_alert_on_cancel"`
	FieldTradesLimit    string `json:"field_trades_limit"`
	FieldMaxPositionUSD string `json:"field_max_position_usd"`
	FieldSlippagePct    string `json:"field_slippage_pct"`
	FieldNegRisk        string `json:"field_neg_risk"`
	FieldSizeMode       string `json:"field_size_mode"`
	FieldBotToken       string `json:"field_bot_token"`
	FieldChatID         string `json:"field_chat_id"`
	FieldAdminChatID    string `json:"field_admin_chat_id"`
	FieldDBPath         string `json:"field_db_path"`
	FieldLogLevel       string `json:"field_log_level"`
	FieldLogFormat      string `json:"field_log_format"`
	FieldTrackPositions string `json:"field_track_positions"`

	// Settings tooltips
	TooltipLanguage        string `json:"tooltip_language"`
	TooltipPrivKey         string `json:"tooltip_priv_key"`
	TooltipAPIKey          string `json:"tooltip_api_key"`
	TooltipAPISecret       string `json:"tooltip_api_secret"`
	TooltipPassphrase      string `json:"tooltip_passphrase"`
	TooltipChainID         string `json:"tooltip_chain_id"`
	TooltipTimeout         string `json:"tooltip_timeout"`
	TooltipMaxRetries      string `json:"tooltip_max_retries"`
	TooltipMonitorEnabled  string `json:"tooltip_monitor_enabled"`
	TooltipMonitorPoll     string `json:"tooltip_monitor_poll"`
	TooltipTradesEnabled   string `json:"tooltip_trades_enabled"`
	TooltipTradesPoll      string `json:"tooltip_trades_poll"`
	TooltipAlertOnFill     string `json:"tooltip_alert_on_fill"`
	TooltipAlertOnCancel   string `json:"tooltip_alert_on_cancel"`
	TooltipTradesLimit     string `json:"tooltip_trades_limit"`
	TooltipTradesTrack     string `json:"tooltip_trades_track"`
	TooltipTradingEnabled  string `json:"tooltip_trading_enabled"`
	TooltipMaxPosition     string `json:"tooltip_max_position"`
	TooltipSlippage        string `json:"tooltip_slippage"`
	TooltipNegRisk         string `json:"tooltip_neg_risk"`
	TooltipCopyEnabled     string `json:"tooltip_copy_enabled"`
	TooltipCopyPoll        string `json:"tooltip_copy_poll"`
	TooltipSizeMode        string `json:"tooltip_size_mode"`
	TooltipTelegramEnabled string `json:"tooltip_telegram_enabled"`
	TooltipBotToken        string `json:"tooltip_bot_token"`
	TooltipChatID          string `json:"tooltip_chat_id"`
	TooltipAdminChatID     string `json:"tooltip_admin_chat_id"`
	TooltipDBEnabled       string `json:"tooltip_db_enabled"`
	TooltipDBPath          string `json:"tooltip_db_path"`
	TooltipLogLevel        string `json:"tooltip_log_level"`
	TooltipLogFormat       string `json:"tooltip_log_format"`

	// Settings UI strings
	SettingsOptions  string `json:"settings_options"`
	SettingsValue    string `json:"settings_value"`
	SettingsUnsaved  string `json:"settings_unsaved"`
	SettingsErrField string `json:"settings_err_field"` // "Error in field «%s»: %v"
	SettingsErrSave  string `json:"settings_err_save"`  // "Save error: %v"

	// Settings help bar
	HelpField      string `json:"help_field"`
	HelpSection    string `json:"help_section"`
	HelpToggle     string `json:"help_toggle"`
	HelpNextOption string `json:"help_next_option"`
	HelpEdit       string `json:"help_edit"`
	HelpSave       string `json:"help_save"`
	HelpReset      string `json:"help_reset"`

	// Wizard
	WizardTitle      string `json:"wizard_title"`
	WizardProgress   string `json:"wizard_progress"`   // "Step %d/%d: %s"
	WizardEmptyField string `json:"wizard_empty_field"`
	WizardWriteError string `json:"wizard_write_error"` // "Config write error: %v"
	WizardContinue   string `json:"wizard_continue"`

	// Wizard step labels
	WizardStep1Label string `json:"wizard_step1_label"`
	WizardStep2Label string `json:"wizard_step2_label"`
	WizardStep3Label string `json:"wizard_step3_label"`
	WizardStep4Label string `json:"wizard_step4_label"`

	// Wizard step hints
	WizardStep1Hint string `json:"wizard_step1_hint"`
	WizardStep2Hint string `json:"wizard_step2_hint"`
	WizardStep3Hint string `json:"wizard_step3_hint"`
	WizardStep4Hint string `json:"wizard_step4_hint"`

	// Log messages (zerolog)
	LogBotStarting          string `json:"log_bot_starting"`
	LogL1Initialized        string `json:"log_l1_initialized"`
	LogTelegramEnabled      string `json:"log_telegram_enabled"`
	LogDatabaseOpened       string `json:"log_database_opened"`
	LogTradesMonitorSkip    string `json:"log_trades_monitor_skip"`
	LogTradesMonitorEnabled string `json:"log_trades_monitor_enabled"`
	LogCopytradingSkipL2    string `json:"log_copytrading_skip_l2"`
	LogCopytradingSkipDB    string `json:"log_copytrading_skip_db"`
	LogCopytradingEnabled   string `json:"log_copytrading_enabled"`
	LogShutdownSignal       string `json:"log_shutdown_signal"`
	LogShuttingDown         string `json:"log_shutting_down"`
	LogBotRunning           string `json:"log_bot_running"`
	LogFatalError           string `json:"log_fatal_error"`
	LogBye                  string `json:"log_bye"`
	LogWSUserEvent          string `json:"log_ws_user_event"`
	LogMonitorStarted       string `json:"log_monitor_started"`
	LogTradesMonitorStarted string `json:"log_trades_monitor_started"`
	LogFailedFetchOrders    string `json:"log_failed_fetch_orders"`
	LogOrdersUpdated        string `json:"log_orders_updated"`
	LogOrderClosed          string `json:"log_order_closed"`
	LogFailedSendAlert      string `json:"log_failed_send_alert"`
	LogNewOrderDetected     string `json:"log_new_order_detected"`
	LogFailedFetchTrades    string `json:"log_failed_fetch_trades"`
	LogTradeExecuted        string `json:"log_trade_executed"`

	// Telegram alert messages
	TgOrderClosed   string `json:"tg_order_closed"`   // "🔔 Order closed: %s"
	TgNewOrder      string `json:"tg_new_order"`      // "📋 New order: %s %s @ %s (size: %s)"
	TgTradeExecuted string `json:"tg_trade_executed"` // "✅ Trade executed: %s %s @ %s (size: %s)"
}
