import 'package:get/get.dart';
import 'package:monolith/app/data/model/model.dart';
import 'package:monolith/app/share/utils.dart';
// created by `dart pub run build_runner build`
import 'package:monolith/objectbox.g.dart';

class StorageService extends GetxService {
  final Store store;
  final Box<PlanetConfig> planet;
  StorageService({
    required this.store,
    required this.planet,
  });

  static Future<StorageService> create() {
    return openStore().then((store) => StorageService(
          store: store,
          planet: store.box<PlanetConfig>(),
        ));
  }

  PlanetConfig? getxPlanetOrFromId(int? id) =>
      ShareUtils.argOrParam(Get.arguments as PlanetConfig?, id, planet.get);

  PlanetConfig? getxPlanet() => ShareUtils.getxArgOrIdParam(planet.get);
}
