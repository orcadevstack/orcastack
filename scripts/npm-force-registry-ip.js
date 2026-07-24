const dns = require('dns');

const registryHosts = new Map([
  ['registry.npmjs.org', '104.16.5.34'],
  ['registry.npmjs.org.', '104.16.5.34'],
]);

const originalLookup = dns.lookup;

dns.lookup = function patchedLookup(hostname, options, callback) {
  const forcedIp = registryHosts.get(hostname);
  if (!forcedIp) {
    return originalLookup.call(dns, hostname, options, callback);
  }

  let resolvedOptions = options;
  let resolvedCallback = callback;

  if (typeof resolvedOptions === 'function') {
    resolvedCallback = resolvedOptions;
    resolvedOptions = {};
  }

  if (resolvedOptions && resolvedOptions.all) {
    return process.nextTick(() => resolvedCallback(null, [{ address: forcedIp, family: 4 }]));
  }

  return process.nextTick(() => resolvedCallback(null, forcedIp, 4));
};