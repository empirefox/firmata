import 'package:get/get.dart';

import '../controllers/planets_controller.dart';

class PlanetsBinding extends Bindings {
  @override
  void dependencies() {
    Get.lazyPut<PlanetsController>(
      () => PlanetsController(),
    );
  }
}
