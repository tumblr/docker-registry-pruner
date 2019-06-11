#!/usr/bin/env ruby
# $0 [--ignore-directories=node_modules,vendor] [path]
# path defaults to the root of your repo
#
require 'find'
require 'pathname'
require 'yaml'
require 'json'
require 'optparse'
$stdout.sync = true

@opts = {
  :yaml_suffixes => %w[
    .yaml
    .yml
  ],
  :json_suffixes => %w[
    .json
  ],
  :ignored_directories => %w[
    /.git/
    /vendor/
    /.gopath~/
    /node_modules/
  ],
  :path => nil,
}

OptionParser.new do |opt|
  opt.on('--ignore-directories DIR[,...]', Array,"ignore additional directories' contents during validation") do |v|
    v.each do |x|
      @opts[:ignored_directories] << File.path(x)
    end
  end
end.parse!

if !ARGV[0].nil?
  @opts[:path] = Pathname.new(ARGV[0]).cleanpath
else
  @opts[:path] = Pathname.new(File.join(File.dirname($0), '..')).cleanpath
end

results = Find.find(@opts[:path]).map do |path|
  success = true
  next if @opts[:ignored_directories].any? {|d| path.include? d }
  if @opts[:yaml_suffixes].any? {|m| path.end_with? m }
    printf "Validating YAML #{path}: "
    begin
      YAML.parse(File.read(path))
      puts "OK"
    rescue => e
      success = false
      puts "ERROR"
      $stderr.puts "YAML validation of #{path} failed: #{e.message}"
    end
  end
  if @opts[:json_suffixes].any? {|m| path.end_with? m }
    printf "Validating JSON #{path}: "
    begin
      JSON.parse(File.open(path).read)
      puts "OK"
    rescue => e
      success = false
      puts "ERROR"
      $stderr.puts "JSON validation of #{path} failed: #{e.message}"
    end
  end
  success
end

exit(1) if results.compact.any? {|r| r == false }
