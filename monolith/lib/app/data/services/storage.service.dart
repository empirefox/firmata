import 'package:get/get.dart';
import 'package:monolith/app/data/model/model.dart';
import 'package:monolith/app/share/utils.dart';
// created by `dart pub run build_runner build`
import 'package:monolith/objectbox.g.dart';

class StorageService extends GetxService {
  final Store _store;
  final Box<PlanetConfig> _planet;
  StorageService({
    required Store store,
    required Box<PlanetConfig> planet,
  })  : _store = store,
        _planet = planet;

  static Future<StorageService> create() {
    return openStore().then((store) => StorageService(
          store: store,
          planet: store.box<PlanetConfig>(),
        ));
  }

  PlanetConfig? getPlanet(int id) {
    if (id == 0) return null;
    return _planet.get(id);
  }

  List<PlanetConfig> getAllPlanet() {
    return _planet.getAll();
  }

  Future<int> setPlanet(PlanetConfig p) {
    return _planet.putAsync(p);
  }

  bool removePlanet(int id) {
    return _planet.remove(id);
  }

  PlanetConfig? getxPlanetOrFromId(int? id) =>
      ShareUtils.argOrParam(Get.arguments as PlanetConfig?, id, getPlanet);

  PlanetConfig? getxPlanet() => ShareUtils.getxArgOrIdParam(getPlanet);
}
