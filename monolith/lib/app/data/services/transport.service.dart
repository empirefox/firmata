import 'package:get/get.dart';
import 'package:grpc/grpc.dart';
import 'package:monolith/app/data/model/model.dart';
import 'package:monolith/app/data/providers/transport.dart';
import 'package:monolith/app/data/services/storage.service.dart';

class TransportService extends GetxService {
  final _ss = Get.find<StorageService>();
  final _transports = <int, Future<Transport>>{};

  bool isOnline(int id) {
    return _transports.containsKey(id);
  }

  Future<Transport>? getOrNull(int id) {
    return _transports[id];
  }

  Future<Transport>? getOrCreateById(int? id,
      {Iterable<ClientInterceptor>? interceptors}) {
    if (id == null) return null;
    var t = _transports[id];
    if (t == null) {
      final config = _ss.getPlanet(id);
      if (config == null) {
        return null;
      }
      return getOrCreate(config);
    }
    return t;
  }

  Future<Transport> getOrCreate(PlanetConfig config,
      {Iterable<ClientInterceptor>? interceptors}) {
    var id = config.id;
    var t = _transports[id];
    if (t == null) {
      Future<void> cb(Transport t) async {
        if ((await _transports[id]) == t) _transports.remove(id);
      }

      final s = interceptors?.toList() ?? [];
      t = Transport.create(
        config,
        onAboutToClose: cb,
      );
      _transports[id] = t;
    }
    return t;
  }

  Future<void> shutdown(int id) async =>
      (await _transports.remove(id))?.shutdown();

  Future<void> terminate(int id) async =>
      (await _transports.remove(id))?.terminate();
}
