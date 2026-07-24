require 'json'
require 'stringio'
require 'webrick'

require_relative './config/application'
require_relative './app/controllers/api/v1/base_controller'
require_relative './app/controllers/api/v1/system_controller'
require_relative './app/controllers/api/v1/auth_controller'
require_relative './app/controllers/api/v1/users_controller'
require_relative './app/controllers/api/v1/projects_controller'
require_relative './app/controllers/api/v1/billing_controller'
require_relative './app/controllers/api/v1/wiki_controller'
require_relative './app/controllers/api/v1/admin_controller'
require_relative './app/controllers/api/v1/webhooks_controller'
require_relative './app/controllers/api/v1/git_controller'
require_relative './app/jobs/webhook_delivery_job'
require_relative './config/routes'

module OrcaStack
  module RubyApp
    class Request
      attr_reader :method, :path, :headers, :body

      def initialize(method:, path:, headers:, body: '')
        @method = method
        @path = path
        @headers = headers
        @body = body
      end
    end

    class Server
      class << self
        def app
          controllers = {
            system: Api::V1::SystemController.new,
            auth: Api::V1::AuthController.new,
            users: Api::V1::UsersController.new,
            projects: Api::V1::ProjectsController.new,
            billing: Api::V1::BillingController.new,
            wiki: Api::V1::WikiController.new,
            admin: Api::V1::AdminController.new,
            webhooks: Api::V1::WebhooksController.new,
            git: Api::V1::GitController.new
          }
          @app ||= Routes.new(controllers)
        end

        def start
          Store.bootstrap!
          port = Integer(ENV.fetch('PORT', '3000'))
          server = WEBrick::HTTPServer.new(Port: port, BindAddress: '0.0.0.0', AccessLog: [], Logger: WEBrick::Log.new($stdout))
          trap('INT') { server.shutdown }
          trap('TERM') { server.shutdown }

          server.mount_proc('/') do |req, res|
            status, payload, extra_headers = call(Request.new(method: req.request_method, path: req.path, headers: req.header, body: req.body.to_s))
            res.status = status
            (extra_headers || {}).each { |key, value| res[key] = value }
            unless extra_headers&.key?('Content-Type')
              res['Content-Type'] = 'application/json'
              payload = JSON.generate(payload)
            end
            res.body = payload
          end

          puts "ruby-app listening on :#{port}"
          server.start
        end

        def rack_call(env)
          body = env['rack.input'] ? env['rack.input'].read : ''
          status, payload, headers = call(Request.new(method: env['REQUEST_METHOD'], path: env['PATH_INFO'], headers: env, body: body))
          headers ||= {}
          unless headers.key?('Content-Type')
            headers['Content-Type'] = 'application/json'
            payload = JSON.generate(payload)
          end
          [status, headers, [payload]]
        end

        def call(request)
          Store.increment_requests!
          handler = app.resolve(request.method, request.path)
          return [404, { error: 'route not found', path: request.path }, nil] unless handler

          result = handler.call(request)
          status, payload, headers = normalize(result)
          [status, payload, headers]
        rescue StandardError => e
          [500, { error: e.message, class: e.class.name }, nil]
        end

        def normalize(result)
          if result.is_a?(Array) && result.length == 3
            result
          elsif result.is_a?(Array) && result.length == 2
            [result[0], result[1], nil]
          else
            [200, result, nil]
          end
        end
      end
    end
  end
end

OrcaStack::RubyApp::Server.start if $PROGRAM_NAME == __FILE__
