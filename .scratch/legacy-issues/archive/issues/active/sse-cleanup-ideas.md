From a Discord user:

> You may want to add a heartbeat helper. Not sure if you're doing that, but I have a go-loop ticking every X seconds pinging all clients, updating their "last connected" status signal and ensuring connections are still around (and stay alive!), else they are removed from the connections store.
> The way I did it in my demo/test repo, data-on-load runs inline js to create a random client-id, then on data-signal-on-patch I ensure to check that's set and that's when the connection starts:
> https://codeberg.org/cljou/datacar/src/branch/master/src/car/handlers/common.clj#L63

> I have not tested this with more than 1 user (only various browser windows), so you may be well ahead of me... shooting a dart and seeing if any of this sticks. Good Luck!

> Edit: Some more details. Just realized /tester doesn't explain what that is- that's the long-running sse connection endpoint (need to rename it). Also, then $clientId signal gets sent with every request and we use the unique client id and store as metadata in connections.clj (manages connections).

I do like the idea of a signal with the client id, though that might introduce security issues--like, would other clients be able to fusk their way into seeing/hitting the keepalive for another client? I honestly don't know how Datastar holds a connection in memory or whatever.

Note that we don't have to use any of these examples in full or even at all! Just ideas!

Another user:

> This is my cloudflare datastar sse connection & messaging component with instructions to keep claude code from messing with it. 
> This is summary code that I had cursor generate from my durable object, hopefully there is enough in here to understand the idea.
> #[durable_object]
> pub struct YourDurableObject {
>     state: State,
>     env: Env,
>     channels: Channels,
>     initialized: bool,
> }
> 
> #[durable_object]
> impl DurableObject for YourDurableObject {
>     fn new(state: State, env: Env) -> Self {
>         Self {
>             state,
>             env,
>             channels: Channels::new(),
>             initialized: false,
>         }
>     }
> 
>     async fn fetch(&mut self, req: Request) -> Result<Response> {
>         self.initialize(Some(&req)).await?;
>         
>         let path = req.path();
>         let method = req.method();
>         let path_parts: Vec<&str> = path.trim_start_matches('/').split('/').collect();
>         let action = path_parts.last().unwrap_or(&"");
> 
>         match (method.as_ref(), *action) {
>             ("GET", "data") => self.handle_sse(req).await,
>             _ => Response::error("Not Found", 404),
>         }
>     }
> 
>     async fn alarm(&mut self) -> Result<Response> {
>         self.initialize(None).await?;
>         
>         if self.channels.connections.is_empty() {
>             return Response::ok("no connections");
>         }
> 
>         let _ = self.broadcast_update().await;
>         
>         if !self.channels.connections.is_empty() {
>             let _ = self.state.storage()
>                 .set_alarm(std::time::Duration::from_secs(5))
>                 .await;
>         }
>         
>         Response::ok("alarm complete")
>     }
> }

>  # DataStar Module: Real-Time Web Infrastructure
> //!
> //! Complete solution for real-time web applications combining DataStar event building
> //! with SSE connection management for Durable Objects.
> //!
> //! ## Key Components
> //!
> //! ### DataStar Event Building
> //! - `Builder`: Constructs properly formatted DataStar SSE events
> //! - `EventType`: All supported DataStar event types (merge fragments, signals, etc.)
> //! - Various `Options` structs for fine-grained control
> //!
> //! ### SSE Infrastructure  
> //! - `Channels`: Reusable connection management for any Durable Object
> //! - `Channel`: Individual SSE connection with proper lifecycle management
> //! - `Broadcast`: Structured result of broadcast operations with cleanup info
> //! - `setup()`: Helper for creating SSE response streams
> //!
> //! ## Usage Pattern for Durable Objects
> //!
> //! ```rust,ignore
> //! pub struct ExampleDurableObject {
> //!     channels: Channels,  // Replace HashMap<String, Channel>
> //! }
> //!
> //! impl ExampleDurableObject {
> //!     async fn broadcast_update(&mut self) {
> //!         let message = datastar::Builder::new()
> //!             .merge_fragments(&self.render_data())
> //!             .to_string();
> //!         
> //!         // Clean separation: channels handles transport
> //!         let result = self.channels.broadcast(&message);
> //!         
> //!         // Business logic handles domain concerns
> //!         if !result.failed_ids.is_empty() {
> //!             self.handle_disconnected_users(&result.failed_ids);
> //!         }
> //!     }
> //! }
> //! ```
> //!
> //! ## Benefits
> //!
> //! - **Reusable**: Any DO can use `Channels` without reimplementing connection logic
> //! - **Maintainable**: Single source of truth for DataStar + SSE patterns
> //! - **Testable**: Clean separation allows independent testing of transport vs business logic
> //! - **Coherent**: DataStar events and their transport mechanism logically unified
> 
> #![allow(dead_code)]
> 
> use serde::Serialize;
> use worker::*;
> use std::fmt::{self, Display};
> use std::collections::HashMap;
> use futures::channel::mpsc::{channel, Sender, TrySendError};
> 
> const SSE_CHANNEL_BUFFER_SIZE: usize = 10;
> 
> #[derive(Debug, Clone, Copy, PartialEq, Eq)]
> pub enum EventType {    
>     MergeFragments,
>     MergeSignals,
>     RemoveFragments,
>     RemoveSignals,
>     ExecuteScript,
>     Redirect,
>     Console,
> }
> 
> impl Display for EventType {
>     fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
>         let event_str = match self {
>             EventType::MergeFragments => "datastar-merge-fragments",
>             EventType::MergeSignals => "datastar-merge-signals",
>             EventType::RemoveFragments => "datastar-remove-fragments",
>             EventType::RemoveSignals => "datastar-remove-signals",
>             EventType::ExecuteScript => "datastar-execute-script",
>             EventType::Redirect => "datastar-redirect",
>             EventType::Console => "datastar-console",
>         };
>         write!(f, "{}", event_str)
>     }
> }
> 
> pub struct FragmentOptions {
>     pub merge_mode: Option<String>,
>     pub selector: Option<String>,
>     pub settle_duration_ms: Option<u32>,
>     pub use_view_transition: bool,
> }
> 
> impl Default for FragmentOptions {
>     fn default() -> Self {
>         Self {
>             merge_mode: None,
>             selector: None,
>             settle_duration_ms: None,
>             use_view_transition: true,
>         }
>     }
> }
> 
> pub struct SignalOptions {
>     pub settle_duration_ms: Option<u32>,
> }
> 
> impl Default for SignalOptions {
>     fn default() -> Self {
>         Self {
>             settle_duration_ms: None,
>         }
>     }
> }
> 
> pub struct RemoveOptions {
>     pub settle_duration_ms: Option<u32>,
>     pub use_view_transition: bool,
> }
> 
> impl Default for RemoveOptions {
>     fn default() -> Self {
>         Self {
>             settle_duration_ms: None,
>             use_view_transition: true,
>         }
>     }
> }
> 
> pub struct RedirectOptions {
>     pub delay_ms: Option<u32>,
> }
> 
> impl Default for RedirectOptions {
>     fn default() -> Self {
>         Self { delay_ms: None }
>     }
> }
> 
> pub struct ConsoleOptions {
>     pub level: Option<String>,
> }
> 
> impl Default for ConsoleOptions {
>     fn default() -> Self {
>         Self { level: None }
>     }
> }
> 
> pub struct Builder {
>     event_lines: Vec<String>,
> }
> 
> // Event builders for the fluent API
> pub struct MergeFragments<'a> {
>     builder: &'a mut Builder,
>     html: String,
>     options: FragmentOptions,
> }
> 
> pub struct MergeSignals<'a, T: Serialize> {
>     builder: &'a mut Builder,
>     signals: &'a T,
>     options: SignalOptions,
> }
> 
> impl Builder {
>     pub fn new() -> Self {
>         Self {
>             event_lines: Vec::new(),
>         }
>     }
> 
>     fn format_multiline_data(&mut self, prefix: &str, content: &str) {
>         if prefix == "fragments" {
>             self.event_lines.extend(
>                 content.lines()
>                     .map(|line| format!("data: {} {}", prefix, line))
>             );
>             return;
>         }
>         
>         let mut lines = content.lines();
>         if let Some(first) = lines.next() {
>             self.event_lines.push(format!("data: {} {}", prefix, first));
>             self.event_lines.extend(lines.map(|line| format!("data: {}", line)));
>         }
>     }
> 
>     fn add_option_line(&mut self, key: &str, value: &str) {
>         self.event_lines.push(format!("data: {} {}", key, value));
>     }
> 
>     // New fluent API methods
>     pub fn merge_fragments(&mut self, html: &str) -> MergeFragments {
>         MergeFragments {
>             builder: self,
>             html: html.to_string(),
>             options: FragmentOptions::default(),
>         }
>     }
>     
>     pub fn merge_signals<'a, T: Serialize>(&'a mut self, signals: &'a T) -> MergeSignals<'a, T> {
>         MergeSignals {
>             builder: self,
>             signals,
>             options: SignalOptions::default(),
>         }
>     }
> 
>     fn add_signals_with_options<T: Serialize>(&mut self, signals: &T, options: &SignalOptions) -> &mut Self {
>         self.event_lines.push(format!("event: {}", EventType::MergeSignals));
>         
>         if let Some(duration) = options.settle_duration_ms {
>             self.add_option_line("settleDuration", &duration.to_string());
>         }
>         
>         let json = serde_json::to_string(signals).unwrap_or_default();
>         self.format_multiline_data("signals", &json);
>         self.event_lines.push(String::new());
>         self
>     }
> 
>     fn add_fragments_with_options(
>         &mut self, 
>         html: &str, 
>         options: &FragmentOptions
>     ) -> &mut Self {
>         self.event_lines.push(format!("event: {}", EventType::MergeFragments));
>         
>         if let Some(mode) = &options.merge_mode {
>             self.add_option_line("mergeMode", mode);
>         }
>         if let Some(sel) = &options.selector {
>             self.add_option_line("selector", sel);
>         }
>         if let Some(duration) = options.settle_duration_ms {
>             self.add_option_line("settleDuration", &duration.to_string());
>         }
>         self.add_option_line("useViewTransition", &options.use_view_transition.to_string());
>         self.event_lines.extend(html.lines().map(|line| format!("data: fragments {}", line)));
>         self.event_lines.push(String::new());
>         self
>     }
>     
> 
>     pub fn remove_fragments(&mut self, selector: &str) -> &mut Self {
>         self.remove_fragments_with_options(selector, &RemoveOptions::default())
>     }
>     
>     pub fn remove_fragments_with_options(&mut self, selector: &str, options: &RemoveOptions) -> &mut Self {
>         self.event_lines.push(format!("event: {}", EventType::RemoveFragments));
>         self.add_option_line("selector", selector);
>         
>         if let Some(duration) = options.settle_duration_ms {
>             self.add_option_line("settleDuration", &duration.to_string());
>         }
>         self.add_option_line("useViewTransition", &options.use_view_transition.to_string());
>         self.event_lines.push(String::new());
>         self
>     }
>     
> 
>     pub fn remove_signals(&mut self, path: &str) -> &mut Self {
>         self.event_lines.push(format!("event: {}", EventType::RemoveSignals));
>         self.event_lines.push(format!("data: paths {}", path));
>         self.event_lines.push(String::new());
>         self
>     }
>     
>     pub fn remove_signals_multi(&mut self, paths: &[&str]) -> &mut Self {
>         self.event_lines.push(format!("event: {}", EventType::RemoveSignals));
>         self.event_lines.extend(paths.iter().map(|path| format!("data: paths {}", path)));
>         self.event_lines.push(String::new());
>         self
>     }
> 
>     pub fn execute_script(&mut self, script: &str) -> &mut Self {
>         self.event_lines.push(format!("event: {}", EventType::ExecuteScript));
>         self.format_multiline_data("script", script);
>         self.event_lines.push(String::new());
>         self
>     }
>     
>     pub fn redirect(&mut self, url: &str) -> &mut Self {
>         self.redirect_with_options(url, &RedirectOptions::default())
>     }
>     
>     pub fn redirect_with_options(&mut self, url: &str, options: &RedirectOptions) -> &mut Self {
>         self.event_lines.push(format!("event: {}", EventType::Redirect));
>         
>         if let Some(delay) = options.delay_ms {
>             self.add_option_line("delay", &delay.to_string());
>         }
>         
>         self.event_lines.push(format!("data: {{\"url\": \"{}\"}}", url));
>         self.event_lines.push(String::new());
>         self
>     }
>     
>     pub fn console_log(&mut self, message: &str) -> &mut Self {
>         self.console_log_with_options(message, &ConsoleOptions::default())
>     }
>     
>     pub fn console_log_with_options(&mut self, message: &str, options: &ConsoleOptions) -> &mut Self {
>         self.event_lines.push(format!("event: {}", EventType::Console));
>         
>         if let Some(level) = &options.level {
>             self.event_lines.push(format!("data: level {}", level));
>         }
>         
>         self.format_multiline_data("message", message);
>         self.event_lines.push(String::new());
>         self
>     }
>     
>     pub fn console_warn(&mut self, message: &str) -> &mut Self {
>         self.console_log_with_options(message, &ConsoleOptions { level: Some("warn".to_string()) })
>     }
>     
>     pub fn console_error(&mut self, message: &str) -> &mut Self {
>         self.console_log_with_options(message, &ConsoleOptions { level: Some("error".to_string()) })
>     }
> 
>     // === OUTPUT METHODS ===
> 
>     pub fn to_string(&self) -> String {
>         self.event_lines.join("\n") + "\n"
>     }
> 
>     pub fn build(self) -> worker::Result<Response> {
>         let mut response = Response::ok(self.to_string())?;
>         let headers = response.headers_mut();
>         headers.set("Content-Type", "text/event-stream")?;
>         headers.set("Cache-Control", "no-cache")?;
>         
>         Ok(response)
>     }
>     
>     pub fn into_response(self) -> Response {
>         let mut response = Response::ok(self.to_string()).unwrap();
>         let headers = response.headers_mut();
>         headers.set("Content-Type", "text/event-stream").unwrap();
>         headers.set("Cache-Control", "no-cache").unwrap();
>         headers.set("Connection", "keep-alive").unwrap();
>         
>         response
>     }
> }
> 
> impl Default for Builder {
>     fn default() -> Self {
>         Self::new()
>     }
> }
> 
> impl Display for Builder {
>     fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
>         write!(f, "{}\n", self.event_lines.join("\n"))
>     }
> }
> 
> // Implementation for MergeFragments builder
> impl<'a> MergeFragments<'a> {
>     pub fn selector(mut self, selector: &str) -> Self {
>         self.options.selector = Some(selector.to_string());
>         self
>     }
>     
>     pub fn merge_mode(mut self, mode: &str) -> Self {
>         self.options.merge_mode = Some(mode.to_string());
>         self
>     }
>     
>     pub fn settle_duration(mut self, duration_ms: u32) -> Self {
>         self.options.settle_duration_ms = Some(duration_ms);
>         self
>     }
>     
>     pub fn use_view_transition(mut self, use_transition: bool) -> Self {
>         self.options.use_view_transition = use_transition;
>         self
>     }
> }
> 
> impl<'a> Drop for MergeFragments<'a> {
>     fn drop(&mut self) {
>         self.builder.add_fragments_with_options(&self.html, &self.options);
>     }
> }
> 
> // Implementation for MergeSignals builder  
> impl<'a, T: Serialize> MergeSignals<'a, T> {
>     pub fn settle_duration(mut self, duration_ms: u32) -> Self {
>         self.options.settle_duration_ms = Some(duration_ms);
>         self
>     }
> }
> 
> impl<'a, T: Serialize> Drop for MergeSignals<'a, T> {
>     fn drop(&mut self) {
>         self.builder.add_signals_with_options(self.signals, &self.options);
>     }
> }
> 
> // ================================================================================================
> // === SSE INFRASTRUCTURE ===
> // ================================================================================================
> //
> // This section provides the transport layer for DataStar events. The Manager pattern
> // allows any Durable Object to add real-time capabilities without reimplementing connection
> // management, broadcasting, and cleanup logic.
> 
> pub fn setup<S>(stream: S) -> worker::Result<Response> 
> where 
>     S: futures::Stream<Item = worker::Result<String>> + 'static,
> {
>     let mut response = Response::from_stream(stream)?;
>     let headers = response.headers_mut();
>     headers.set("Content-Type", "text/event-stream")?;
>     headers.set("Cache-Control", "no-cache")?;
>     headers.set("Connection", "keep-alive")?;
>     headers.set("X-Content-Type-Options", "nosniff")?;
>     Ok(response)
> }
> 
> pub struct Broadcast {
>     pub sent_count: usize,
>     pub failed_ids: Vec<String>,
> }
> 
> pub struct Channels {
>     pub connections: HashMap<String, Channel>,
> }
> 
> impl Channels {
>     pub fn new() -> Self {
>         Self {
>             connections: HashMap::new(),
>         }
>     }
> 
>     pub fn add_connection(&mut self, id: String, channel: Channel) -> worker::Result<()> {
>         self.connections.insert(id, channel);
>         Ok(())
>     }
> 
>     pub fn send(&mut self, id: &str, message: &str) -> Result<()> {
>         if let Some(channel) = self.connections.get_mut(id) {
>             channel.send(message)
>                 .map_err(|e| {
>                     console_log!("🔌 Channel {} disconnected: {:?}", id, e);
>                     Error::RustError(format!("Channel {} disconnected", id))
>                 })?;
>             Ok(())
>         } else {
>             Err(Error::RustError(format!("Channel {} not found", id)))
>         }
>     }
> 
>     pub fn broadcast(&mut self, message: &str) -> Broadcast {
>         let mut failed_ids = Vec::new();
>         let initial_count = self.connections.len();
> 
>         for (id, channel) in self.connections.iter_mut() {
>             if let Err(e) = channel.send(message) {
>                 console_log!("🔌 Connection {} disconnected during broadcast: {:?}", id, e);
>                 failed_ids.push(id.clone());
>             }
>         }
> 
>         for id in &failed_ids {
>             self.remove_connection(id);
>         }
> 
>         let sent_count = initial_count - failed_ids.len();
> 
>         Broadcast {
>             sent_count,
>             failed_ids,
>         }
>     }
> 
>     pub fn remove_connection(&mut self, connection_id: &str) -> Option<Channel> {
>         if let Some(mut channel) = self.connections.remove(connection_id) {
>             let _ = futures::executor::block_on(channel.close());
>             Some(channel)
>         } else {
>             None
>         }
>     }
> 
> }
> 
> pub struct Channel {
>     pub sender: Sender<worker::Result<String>>,
>     pub closed: bool,
> }
> 
> impl Channel {
> 
>     pub fn send(&mut self, message: &str) -> std::result::Result<(), TrySendError<worker::Result<String>>> {
>         // DON'T check self.closed - causes unwrap_err() bugs that create uncaught errors
>         self.sender.try_send(Ok(message.to_string()))
>     }
>     
>     pub async fn close(&mut self) -> worker::Result<()> {
>         if self.closed {
>             return Ok(());
>         }        
>         self.closed = true;        
>         self.sender.close_channel();        
>         Ok(())
>     }
> 
>     pub fn setup() -> worker::Result<(Self, Response)> {
>         let (sender, receiver) = channel::<worker::Result<String>>(SSE_CHANNEL_BUFFER_SIZE);
>         let channel = Self { sender, closed: false };
>         let response = setup(receiver)?;
>         Ok((channel, response))
>     }
> }

> impl YourDurableObject {
>     async fn initialize(&mut self, req: Option<&Request>) -> Result<()> {
>         if self.initialized {
>             return Ok(());
>         }
>         self.initialized = true;
>         Ok(())
>     }
> 
>     async fn handle_sse(&mut self, req: Request) -> Result<Response> {
>         // Create SSE channel and response
>         let (channel, response) = Channel::setup()?;
>         
>         // Get unique connection ID
>         let connection_id = get_valid_session_data(&req, &self.env).unwrap().user_id;
>         
>         // Store the channel for future broadcasts
>         self.channels.add_connection(connection_id.clone(), channel)?;
>         
>         // Send initial data
>         let html = format!(include_str!("your_template.html"), initial_data);
>         let mut builder = Builder::new();
>         builder.merge_fragments(&html);
>         if let Some(channel) = self.channels.connections.get_mut(&connection_id) {
>             let _ = channel.send(&builder.to_string());
>         }
> 
>         // Broadcast to all other connections
>         let _ = self.broadcast_update().await;
>         
>         Ok(response)
>     }
> 
>     async fn broadcast_update(&mut self) -> Result<()> {
>         let html = format!(include_str!("your_template.html"), updated_data);
>         let mut builder = Builder::new();
>         builder.merge_fragments(&html);
>         let message = builder.to_string();
>         
>         let result = self.channels.broadcast(&message);
>         
>         if !result.failed_ids.is_empty() {
>             // Handle any business logic cleanup for failed connections
>             self.update_state()?;
>             
>             // Broadcast updated state to remaining connections
>             let updated_html = format!(include_str!("your_template.html"), self.current_state);
>             let mut builder = Builder::new();
>             builder.merge_fragments(&updated_html);
>             self.channels.broadcast(&builder.to_string());
>         }
>         
>         Ok(())
>     }
> }
> The key insight: We don't need explicit disconnect detection because:
> Response::from_stream() creates a streaming response
> When client disconnects, the stream fails
> Our next try_send attempt fails
> We detect this in broadcast and clean up
> This pattern is self-maintaining - no polling or explicit keepalive needed. The streaming response handles disconnect detection for us at the transport level.
> The beauty is that it's completely passive - we only find out about dead connections when we try to write to them, which is exactly when we need to know.
> 
> It usually takes 4-5 messages being sent through the SSE broadcast for a dead connection to error and fail out, that was the big difference with websockets which are instantaneous because of the bidirectional aspects.  I use an incremental backoff on the alarm to ensure I'm not overwhelming my worker and exceeding the Cloudflare limits, but I'm doing my math based on 1000 concurrent connections to a single page, 24/7 and it's still only like less than $100 if I exceed the requests per month limits for the $5 worker.
> I plan to clean the craft project up and provide it as a working guide once I'm done, but I need to get this to work on my actual project now that I've prototyped it out.
> Peter Hammond — 6/13/25, 5:54 PM
> The secret sauce is Cloudflare keeping the SSE channels live even when the Durable Object is hibernated, combined with using TrySendError (which Claude discovered during refactoring—it wasn't in any of the demos or SDK docs that I was able to find, which are pretty lean for workers-rs). I was having issues with dead connections before implementing this.
> use futures::channel::mpsc::{channel, Sender, TrySendError};
>  
> Note this would all be a lot simpler if you do it direclty with the native Javascript workers, but I want the RUST benefits.  Hopefully this is helpful. 
> /me braces for @Cult Leader to ban me a heretic and burn me at the stake for using Claude Code.  🔥 🔥 🔥
> But with 
> ● Bash(git commit -m "♻️ refactor(core): restructure codebase with new CF modules and Datastar integration")
>   ⎿  [feature/v0.6.0-update-devops-and-config bb5820d] ♻️ refactor(core): restructure codebase with new CF modules and Datastar integration
>       24 files changed, 2009 insertions(+), 376 deletions(-)
>       create mode 100644 src/cf_datastar.rs
> 
> Worth it.  🙂




Caddy would also probably help, to hold a connection open if upstream is not responding? Right?
