
scApp.controller("CRTCCheck", ["$scope", "$location", "$sc_utility", "$sc_nav", function($scope, $location, $sc_utility, $sc_nav){
    $sc_nav.in_rtc_check();
    $sc_utility.refresh.stop();
}]);
