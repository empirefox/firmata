import 'package:get/get.dart';
import 'package:monolith/app/data/model/model.dart';
import 'package:monolith/app/data/services/storage.service.dart';
import 'package:monolith/app/data/services/transport.service.dart';
import 'package:monolith/app/routes/app_pages.dart';

class PlanetsController extends GetxController {
  final _ss = Get.find<StorageService>();
  final _ts = Get.find<TransportService>();

  late List<PlanetConfig> planets;

  @override
  void onInit() {
    super.onInit();
    planets = _ss.planet.getAll();
  }

  @override
  void onReady() {
    super.onReady();
  }

  @override
  void onClose() {}

  bool isOnline(PlanetConfig planet) => _ts.isOnline(planet.id);

  Future<String>? findServerName(int id) =>
      _ts.getOrNull(id)?.then((t) => t.config.nick);

  void about() => Get.toNamed(Routes.ABOUT);

  void create() =>
      Get.toNamed<PlanetConfig>('${Routes.PLANET}?id=0')?.then((n) {
        planets.insert(0, n!);
        update([n.id]);
      });

  void delete(PlanetConfig planet) => _ss.planet.remove(planet.id);

  void edit(PlanetConfig planet) => Get.toNamed<PlanetConfig>(
        '${Routes.PLANET_EDIT}?id=${planet.id}',
        arguments: planet,
      )?.then((n) {
        planets[planets.indexOf(planet)] = n!;
        update([n.id]);
      });

  void shutdown(PlanetConfig planet) =>
      _ts.shutdown(planet.id).then((_) => update([planet.id]));

  void view(PlanetConfig planet) async {
    final isDown = await Get.toNamed<bool>(
      '${Routes.PLANET}?id=${planet.id}',
      arguments: planet,
    );
    if (isDown == true) {
      await _ts.shutdown(planet.id);
      update([planet.id]);
    }
  }
}
