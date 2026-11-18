package validation

import (
	"fmt"
	"strings"
	"testing"
)

func TestAccuracyHunt(t *testing.T) {
	cases := []struct {
		id, argument string
		valid        bool
	}{
		{"H001", "1. If X is red, then X is warm.\n2. Nobody is red.\n3. Therefore, Nobody is warm. [1,2]\n", false},
		{"H002", "1. If X is red, then X is warm.\n2. No one is red.\n3. Therefore, No one is warm. [1,2]\n", false},
		{"H003", "1. If X is red, then X is warm.\n2. Few robots is red.\n3. Therefore, Few robots is warm. [1,2]\n", false},
		{"H004", "1. If X is red, then X is warm.\n2. Most robots is red.\n3. Therefore, Most robots is warm. [1,2]\n", true},
		{"H005", "1. If X is red, then X is warm.\n2. At most two robots is red.\n3. Therefore, At most two robots is warm. [1,2]\n", false},
		{"H006", "1. If X is red, then X is warm.\n2. Exactly two robots is red.\n3. Therefore, Exactly two robots is warm. [1,2]\n", false},
		{"H007", "1. If X is red, then X is warm.\n2. Only Alpha is red.\n3. Therefore, Only Alpha is warm. [1,2]\n", false},
		{"H008", "1. If X is red, then X is warm.\n2. Not every robot is red.\n3. Therefore, Not every robot is warm. [1,2]\n", false},
		{"H009", "1. If X is red, then X is warm.\n2. Neither robot is red.\n3. Therefore, Neither robot is warm. [1,2]\n", false},
		{"H010", "1. If X is red, then X is warm.\n2. All but one robot is red.\n3. Therefore, All but one robot is warm. [1,2]\n", false},
		{"H011", "1. If X is red, then X is warm.\n2. Almost every robot is red.\n3. Therefore, Almost every robot is warm. [1,2]\n", false},
		{"H012", "1. If X is red, then X is warm.\n2. Hardly any robot is red.\n3. Therefore, Hardly any robot is warm. [1,2]\n", false},
		{"H013", "1. If X is red, then X is warm.\n2. Less than half the robots is red.\n3. Therefore, Less than half the robots is warm. [1,2]\n", false},
		{"H014", "1. If X is red, then X is warm.\n2. More than two robots is red.\n3. Therefore, More than two robots is warm. [1,2]\n", true},
		{"H015", "1. If X is red, then X is warm.\n2. The alleged robot is red.\n3. Therefore, The alleged robot is warm. [1,2]\n", true},
		{"H016", "1. If X is red, then X is warm.\n2. The fake robot is red.\n3. Therefore, The fake robot is warm. [1,2]\n", true},
		{"H017", "1. If X is red, then X is warm.\n2. It is alleged that Alpha is red.\n3. Therefore, It is alleged that Alpha is warm. [1,2]\n", false},
		{"H018", "1. If X is red, then X is warm.\n2. It is rumored that Alpha is red.\n3. Therefore, It is rumored that Alpha is warm. [1,2]\n", false},
		{"H019", "1. If X is red, then X is warm.\n2. It is conceivable that Alpha is red.\n3. Therefore, It is conceivable that Alpha is warm. [1,2]\n", false},
		{"H020", "1. If X is red, then X is warm.\n2. It is conceivable that the large red robot is red.\n3. Therefore, It is conceivable that the large red robot is warm. [1,2]\n", false},
		{"H021", "1. If X is red, then X is warm.\n2. The observer doubts whether Alpha is red.\n3. Therefore, The observer doubts whether Alpha is warm. [1,2]\n", false},
		{"H022", "1. If X is red, then X is warm.\n2. The observer imagines that Alpha is red.\n3. Therefore, The observer imagines that Alpha is warm. [1,2]\n", false},
		{"H023", "1. If X is red, then X is warm.\n2. The observer dreams that Alpha is red.\n3. Therefore, The observer dreams that Alpha is warm. [1,2]\n", false},
		{"H024", "1. If X is red, then X is warm.\n2. The observer pretends that Alpha is red.\n3. Therefore, The observer pretends that Alpha is warm. [1,2]\n", false},
		{"H025", "1. If X is red, then X is warm.\n2. The observer believes Alpha is red.\n3. Therefore, The observer believes Alpha is warm. [1,2]\n", false},
		{"H026", "1. If X is red, then X is warm.\n2. The observer thinks Alpha is red.\n3. Therefore, The observer thinks Alpha is warm. [1,2]\n", false},
		{"H027", "1. If X is red, then X is warm.\n2. Possibly Alpha is red.\n3. Therefore, Possibly Alpha is warm. [1,2]\n", false},
		{"H028", "1. If X is red, then X is warm.\n2. Allegedly Alpha is red.\n3. Therefore, Allegedly Alpha is warm. [1,2]\n", false},
		{"H029", "1. If X is red, then X is warm.\n2. Perhaps Alpha is red.\n3. Therefore, Perhaps Alpha is warm. [1,2]\n", false},
		{"H030", "1. If X is red, then X is warm.\n2. Alpha, who is red, is red.\n3. Therefore, Alpha, who is red, is warm. [1,2]\n", true},
		{"H031", "1. If X is red, then X is warm.\n2. The robot that is red is red.\n3. Therefore, The robot that is red is warm. [1,2]\n", true},
		{"H032", "1. If X is red, then X is warm.\n2. Alpha or Beta is red.\n3. Therefore, Alpha or Beta is warm. [1,2]\n", true},
		{"H033", "1. If X is red, then X is warm.\n2. Alpha and Beta is red.\n3. Therefore, Alpha and Beta is warm. [1,2]\n", true},
		{"H034", "1. If X is red, then X is warm.\n2. Every robot is red.\n3. Therefore, Every robot is warm. [1,2]\n", true},
		{"H035", "1. If X is red, then X is warm.\n2. Some robot is red.\n3. Therefore, Some robot is warm. [1,2]\n", true},
		{"H036", "1. If X is red, then X is warm.\n2. A robot is red.\n3. Therefore, A robot is warm. [1,2]\n", true},
		{"H037", "1. Nobody is red.\n2. Therefore, Something is red. [1]\n", false},
		{"H038", "1. No one is red.\n2. Therefore, Something is red. [1]\n", false},
		{"H039", "1. Few robots is red.\n2. Therefore, Something is red. [1]\n", false},
		{"H040", "1. Most robots is red.\n2. Therefore, Something is red. [1]\n", false},
		{"H041", "1. At most two robots is red.\n2. Therefore, Something is red. [1]\n", false},
		{"H042", "1. Exactly two robots is red.\n2. Therefore, Something is red. [1]\n", false},
		{"H043", "1. Only Alpha is red.\n2. Therefore, Something is red. [1]\n", false},
		{"H044", "1. Not every robot is red.\n2. Therefore, Something is red. [1]\n", false},
		{"H045", "1. Neither robot is red.\n2. Therefore, Something is red. [1]\n", false},
		{"H046", "1. All but one robot is red.\n2. Therefore, Something is red. [1]\n", false},
		{"H047", "1. Almost every robot is red.\n2. Therefore, Something is red. [1]\n", true},
		{"H048", "1. Hardly any robot is red.\n2. Therefore, Something is red. [1]\n", false},
		{"H049", "1. Less than half the robots is red.\n2. Therefore, Something is red. [1]\n", false},
		{"H050", "1. More than two robots is red.\n2. Therefore, Something is red. [1]\n", true},
		{"H051", "1. The alleged robot is red.\n2. Therefore, Something is red. [1]\n", true},
		{"H052", "1. The fake robot is red.\n2. Therefore, Something is red. [1]\n", true},
		{"H053", "1. It is alleged that Alpha is red.\n2. Therefore, Something is red. [1]\n", false},
		{"H054", "1. It is rumored that Alpha is red.\n2. Therefore, Something is red. [1]\n", false},
		{"H055", "1. It is conceivable that Alpha is red.\n2. Therefore, Something is red. [1]\n", false},
		{"H056", "1. It is conceivable that the large red robot is red.\n2. Therefore, Something is red. [1]\n", false},
		{"H057", "1. The observer doubts whether Alpha is red.\n2. Therefore, Something is red. [1]\n", false},
		{"H058", "1. The observer imagines that Alpha is red.\n2. Therefore, Something is red. [1]\n", false},
		{"H059", "1. The observer dreams that Alpha is red.\n2. Therefore, Something is red. [1]\n", false},
		{"H060", "1. The observer pretends that Alpha is red.\n2. Therefore, Something is red. [1]\n", false},
		{"H061", "1. The observer believes Alpha is red.\n2. Therefore, Something is red. [1]\n", false},
		{"H062", "1. The observer thinks Alpha is red.\n2. Therefore, Something is red. [1]\n", false},
		{"H063", "1. Possibly Alpha is red.\n2. Therefore, Something is red. [1]\n", false},
		{"H064", "1. Allegedly Alpha is red.\n2. Therefore, Something is red. [1]\n", false},
		{"H065", "1. Perhaps Alpha is red.\n2. Therefore, Something is red. [1]\n", false},
		{"H066", "1. Alpha, who is red, is red.\n2. Therefore, Something is red. [1]\n", false},
		{"H067", "1. The robot that is red is red.\n2. Therefore, Something is red. [1]\n", true},
		{"H068", "1. Alpha or Beta is red.\n2. Therefore, Something is red. [1]\n", true},
		{"H069", "1. Alpha and Beta is red.\n2. Therefore, Something is red. [1]\n", true},
		{"H070", "1. Every robot is red.\n2. Therefore, Something is red. [1]\n", false},
		{"H071", "1. Some robot is red.\n2. Therefore, Something is red. [1]\n", false},
		{"H072", "1. A robot is red.\n2. Therefore, Something is red. [1]\n", false},
		{"H073", "1. Every red thing is warm.\n2. Nobody is a red thing.\n3. Therefore, Nobody is warm. [1,2]\n", false},
		{"H074", "1. Every red thing is warm.\n2. No one is a red thing.\n3. Therefore, No one is warm. [1,2]\n", false},
		{"H075", "1. Every red thing is warm.\n2. Few robots is a red thing.\n3. Therefore, Few robots is warm. [1,2]\n", false},
		{"H076", "1. Every red thing is warm.\n2. Most robots is a red thing.\n3. Therefore, Most robots is warm. [1,2]\n", true},
		{"H077", "1. Every red thing is warm.\n2. At most two robots is a red thing.\n3. Therefore, At most two robots is warm. [1,2]\n", false},
		{"H078", "1. Every red thing is warm.\n2. Exactly two robots is a red thing.\n3. Therefore, Exactly two robots is warm. [1,2]\n", false},
		{"H079", "1. Every red thing is warm.\n2. Only Alpha is a red thing.\n3. Therefore, Only Alpha is warm. [1,2]\n", false},
		{"H080", "1. Every red thing is warm.\n2. Not every robot is a red thing.\n3. Therefore, Not every robot is warm. [1,2]\n", false},
		{"H081", "1. Every red thing is warm.\n2. Neither robot is a red thing.\n3. Therefore, Neither robot is warm. [1,2]\n", false},
		{"H082", "1. Every red thing is warm.\n2. All but one robot is a red thing.\n3. Therefore, All but one robot is warm. [1,2]\n", false},
		{"H083", "1. Every red thing is warm.\n2. Almost every robot is a red thing.\n3. Therefore, Almost every robot is warm. [1,2]\n", false},
		{"H084", "1. Every red thing is warm.\n2. Hardly any robot is a red thing.\n3. Therefore, Hardly any robot is warm. [1,2]\n", false},
		{"H085", "1. Every red thing is warm.\n2. Less than half the robots is a red thing.\n3. Therefore, Less than half the robots is warm. [1,2]\n", false},
		{"H086", "1. Every red thing is warm.\n2. More than two robots is a red thing.\n3. Therefore, More than two robots is warm. [1,2]\n", true},
		{"H087", "1. Every red thing is warm.\n2. The alleged robot is a red thing.\n3. Therefore, The alleged robot is warm. [1,2]\n", true},
		{"H088", "1. Every red thing is warm.\n2. The fake robot is a red thing.\n3. Therefore, The fake robot is warm. [1,2]\n", true},
		{"H089", "1. Every red thing is warm.\n2. It is alleged that Alpha is a red thing.\n3. Therefore, It is alleged that Alpha is warm. [1,2]\n", false},
		{"H090", "1. Every red thing is warm.\n2. It is rumored that Alpha is a red thing.\n3. Therefore, It is rumored that Alpha is warm. [1,2]\n", false},
		{"H091", "1. Every red thing is warm.\n2. It is conceivable that Alpha is a red thing.\n3. Therefore, It is conceivable that Alpha is warm. [1,2]\n", false},
		{"H092", "1. Every red thing is warm.\n2. It is conceivable that the large red robot is a red thing.\n3. Therefore, It is conceivable that the large red robot is warm. [1,2]\n", false},
		{"H093", "1. Every red thing is warm.\n2. The observer doubts whether Alpha is a red thing.\n3. Therefore, The observer doubts whether Alpha is warm. [1,2]\n", false},
		{"H094", "1. Every red thing is warm.\n2. The observer imagines that Alpha is a red thing.\n3. Therefore, The observer imagines that Alpha is warm. [1,2]\n", false},
		{"H095", "1. Every red thing is warm.\n2. The observer dreams that Alpha is a red thing.\n3. Therefore, The observer dreams that Alpha is warm. [1,2]\n", false},
		{"H096", "1. Every red thing is warm.\n2. The observer pretends that Alpha is a red thing.\n3. Therefore, The observer pretends that Alpha is warm. [1,2]\n", false},
		{"H097", "1. The observer believes that Alpha is red and Beta is blue.\n2. Therefore, Beta is blue. [1]\n", false},
		{"H098", "1. The observer believes that Alpha is red or Beta is blue.\n2. It is not the case that The observer believes that Alpha is red.\n3. Therefore, Beta is blue. [1,2]\n", false},
		{"H099", "1. The observer imagines that Alpha is red and Beta is blue.\n2. Therefore, Beta is blue. [1]\n", false},
		{"H100", "1. The observer imagines that Alpha is red or Beta is blue.\n2. It is not the case that The observer imagines that Alpha is red.\n3. Therefore, Beta is blue. [1,2]\n", false},
		{"H101", "1. The observer pretends that Alpha is red and Beta is blue.\n2. Therefore, Beta is blue. [1]\n", false},
		{"H102", "1. The observer pretends that Alpha is red or Beta is blue.\n2. It is not the case that The observer pretends that Alpha is red.\n3. Therefore, Beta is blue. [1,2]\n", false},
		{"H103", "1. The observer dreams that Alpha is red and Beta is blue.\n2. Therefore, Beta is blue. [1]\n", false},
		{"H104", "1. The observer dreams that Alpha is red or Beta is blue.\n2. It is not the case that The observer dreams that Alpha is red.\n3. Therefore, Beta is blue. [1,2]\n", false},
		{"H105", "1. It is alleged that Alpha is red and Beta is blue.\n2. Therefore, Beta is blue. [1]\n", false},
		{"H106", "1. It is alleged that Alpha is red or Beta is blue.\n2. It is not the case that It is alleged that Alpha is red.\n3. Therefore, Beta is blue. [1,2]\n", false},
		{"H107", "1. It is rumored that Alpha is red and Beta is blue.\n2. Therefore, Beta is blue. [1]\n", false},
		{"H108", "1. It is rumored that Alpha is red or Beta is blue.\n2. It is not the case that It is rumored that Alpha is red.\n3. Therefore, Beta is blue. [1,2]\n", false},
		{"H109", "1. It is possible that Alpha is red and Beta is blue.\n2. Therefore, Beta is blue. [1]\n", false},
		{"H110", "1. It is possible that Alpha is red or Beta is blue.\n2. It is not the case that It is possible that Alpha is red.\n3. Therefore, Beta is blue. [1,2]\n", false},
		{"H111", "1. It is likely that Alpha is red and Beta is blue.\n2. Therefore, Beta is blue. [1]\n", false},
		{"H112", "1. It is likely that Alpha is red or Beta is blue.\n2. It is not the case that It is likely that Alpha is red.\n3. Therefore, Beta is blue. [1,2]\n", false},
		{"H113", "1. Perhaps Alpha is red and Beta is blue.\n2. Therefore, Beta is blue. [1]\n", false},
		{"H114", "1. Perhaps Alpha is red or Beta is blue.\n2. It is not the case that Perhaps Alpha is red.\n3. Therefore, Beta is blue. [1,2]\n", false},
		{"H115", "1. Allegedly Alpha is red and Beta is blue.\n2. Therefore, Beta is blue. [1]\n", false},
		{"H116", "1. Allegedly Alpha is red or Beta is blue.\n2. It is not the case that Allegedly Alpha is red.\n3. Therefore, Beta is blue. [1,2]\n", false},
		{"H117", "1. Usually Alpha is red and Beta is blue.\n2. Therefore, Beta is blue. [1]\n", false},
		{"H118", "1. Usually Alpha is red or Beta is blue.\n2. It is not the case that Usually Alpha is red.\n3. Therefore, Beta is blue. [1,2]\n", false},
		{"H119", "1. Sometimes Alpha is red and Beta is blue.\n2. Therefore, Beta is blue. [1]\n", false},
		{"H120", "1. Sometimes Alpha is red or Beta is blue.\n2. It is not the case that Sometimes Alpha is red.\n3. Therefore, Beta is blue. [1,2]\n", false},
		{"H121", "1. Yesterday Alpha is red and Beta is blue.\n2. Therefore, Beta is blue. [1]\n", false},
		{"H122", "1. Yesterday Alpha is red or Beta is blue.\n2. It is not the case that Yesterday Alpha is red.\n3. Therefore, Beta is blue. [1,2]\n", false},
		{"H123", "1. Once Alpha is red and Beta is blue.\n2. Therefore, Beta is blue. [1]\n", false},
		{"H124", "1. Once Alpha is red or Beta is blue.\n2. It is not the case that Once Alpha is red.\n3. Therefore, Beta is blue. [1,2]\n", false},
		{"H125", "1. If the switch is on, then Alpha is red and Beta is blue.\n2. Therefore, Beta is blue. [1]\n", false},
		{"H126", "1. If the switch is on, then Alpha is red or Beta is blue.\n2. It is not the case that If the switch is on, then Alpha is red.\n3. Therefore, Beta is blue. [1,2]\n", false},
		{"H127", "1. It is not the case that Alpha is red and Beta is blue.\n2. Therefore, Beta is blue. [1]\n", false},
		{"H128", "1. It is not the case that Alpha is red or Beta is blue.\n2. It is not the case that It is not the case that Alpha is red.\n3. Therefore, Beta is blue. [1,2]\n", false},
		{"H129", "1. The robots, which are red, are warm.\n2. Therefore, The robots which are red are warm. [1]\n", true},
		{"H130", "1. The robots which are red are warm.\n2. Therefore, The robots, which are red, are warm. [1]\n", true}, // known gap: the FOL path ignores the relative comma (docs/coverage.md)
		{"H131", "1. Alpha eats, shoots, and leaves.\n2. Therefore, Alpha eats shoots and leaves. [1]\n", true},           // known gap: verb-list commas are ignored
		{"H132", "1. Alpha eats shoots and leaves.\n2. Therefore, Alpha eats, shoots, and leaves. [1]\n", true},           // known gap: verb-list commas are ignored
		{"H133", "1. The robot, Alpha, is red.\n2. Therefore, The robot Alpha is red. [1]\n", true},
		{"H134", "1. Let us eat, robots.\n2. Therefore, Let us eat robots. [1]\n", true}, // known gap: vocative comma is ignored
		{"H135", "1. All axes are valuable.\n2. Alpha is a axe.\n3. Therefore, Alpha is valuable. [1,2]\n", false},
		{"H136", "1. All axes are valuable.\n2. Alpha is a axis.\n3. Therefore, Alpha is valuable. [1,2]\n", false},
		{"H137", "1. All clothes are valuable.\n2. Alpha is a cloth.\n3. Therefore, Alpha is valuable. [1,2]\n", false},
		{"H138", "1. All species are valuable.\n2. Alpha is a specie.\n3. Therefore, Alpha is valuable. [1,2]\n", false},
		{"H139", "1. All means are valuable.\n2. Alpha is a mean.\n3. Therefore, Alpha is valuable. [1,2]\n", false},
		{"H140", "1. All premises are valuable.\n2. Alpha is a premise.\n3. Therefore, Alpha is valuable. [1,2]\n", true},
		{"H141", "1. All customs are valuable.\n2. Alpha is a custom.\n3. Therefore, Alpha is valuable. [1,2]\n", false},
		{"H142", "1. All arms are valuable.\n2. Alpha is a arm.\n3. Therefore, Alpha is valuable. [1,2]\n", false},
		{"H143", "1. All glasses are valuable.\n2. Alpha is a glass.\n3. Therefore, Alpha is valuable. [1,2]\n", false},
		{"H144", "1. All goods are valuable.\n2. Alpha is a good.\n3. Therefore, Alpha is valuable. [1,2]\n", false},
		{"H145", "1. All news are valuable.\n2. Alpha is a new.\n3. Therefore, Alpha is valuable. [1,2]\n", false},
		{"H146", "1. Alpha is identical to Beta.\n2. Alpha is allegedly red.\n3. Therefore, Beta is allegedly red. [1,2]\n", false},
		{"H147", "1. Alpha is identical to Beta.\n2. Alpha is almost red.\n3. Therefore, Beta is almost red. [1,2]\n", true},
		{"H148", "1. Alpha is identical to Beta.\n2. Alpha is a fake robot.\n3. Therefore, Beta is a fake robot. [1,2]\n", true},
		{"H149", "1. Alpha is identical to Beta.\n2. The observer believes that Alpha is red.\n3. Therefore, The observer believes that Beta is red. [1,2]\n", false},
		{"H150", "1. Alpha is identical to Beta.\n2. The observer imagines Alpha is red.\n3. Therefore, The observer imagines Beta is red. [1,2]\n", false},
		{"H151", "1. Alpha is identical to Beta.\n2. The observer dreams about Alpha.\n3. Therefore, The observer dreams about Beta. [1,2]\n", false},
		{"H152", "1. Alpha is identical to Beta.\n2. The observer seeks Alpha.\n3. Therefore, The observer seeks Beta. [1,2]\n", false},
		{"H153", "1. Alpha is identical to Beta.\n2. The observer wants Alpha.\n3. Therefore, The observer wants Beta. [1,2]\n", false},
		{"H154", "1. Alpha is identical to Beta.\n2. The observer painted Alpha.\n3. Therefore, The observer painted Beta. [1,2]\n", false}, // v35.0: object position
		{"H155", "1. Alpha is identical to Beta.\n2. Alpha was red.\n3. Therefore, Beta was red. [1,2]\n", true},
		{"H156", "1. Alpha is identical to Beta.\n2. Alpha might be red.\n3. Therefore, Beta might be red. [1,2]\n", true}, // v35.0: a name as the subject of a modal sentence
		{"H157", "1. Alpha is identical to Beta.\n2. Alpha is called red.\n3. Therefore, Beta is called red. [1,2]\n", false},
		{"H158", "1. Alpha is identical to Beta.\n2. Alpha is named red.\n3. Therefore, Beta is named red. [1,2]\n", false},
		{"H159", "1. Alpha is identical to Beta.\n2. Alpha is similar to Gamma.\n3. Therefore, Beta is similar to Gamma. [1,2]\n", true},
		{"H160", "1. Alpha is identical to Beta.\n2. Alpha is near Gamma.\n3. Therefore, Beta is near Gamma. [1,2]\n", true},
		{"H161", "1. Alpha is near Beta.\n2. Beta is near Gamma.\n3. Therefore, Alpha is near Gamma. [1,2]\n", false},
		{"H162", "1. Alpha is similar to Beta.\n2. Beta is similar to Gamma.\n3. Therefore, Alpha is similar to Gamma. [1,2]\n", false},
		{"H163", "1. Alpha is taller than Beta.\n2. Beta is taller than Gamma.\n3. Therefore, Alpha is taller than Gamma. [1,2]\n", false},
		{"H164", "1. Alpha is as tall as Beta.\n2. Beta is as tall as Gamma.\n3. Therefore, Alpha is as tall as Gamma. [1,2]\n", false},
		{"H165", "1. Alpha is adjacent to Beta.\n2. Beta is adjacent to Gamma.\n3. Therefore, Alpha is adjacent to Gamma. [1,2]\n", false},
		{"H166", "1. Alpha is a friend of Beta.\n2. Beta is a friend of Gamma.\n3. Therefore, Alpha is a friend of Gamma. [1,2]\n", false},
		{"H167", "1. Alpha is related to Beta.\n2. Beta is related to Gamma.\n3. Therefore, Alpha is related to Gamma. [1,2]\n", false},
		{"H168", "1. Alpha is older than Beta.\n2. Beta is older than Gamma.\n3. Therefore, Alpha is older than Gamma. [1,2]\n", false},
		{"H169", "1. Alpha is north of Beta.\n2. Beta is north of Gamma.\n3. Therefore, Alpha is north of Gamma. [1,2]\n", false},
		{"H170", "1. If Alpha is red, then Beta is blue.\n2. Alpha is red.\n3. Therefore, Beta is blue. [1,2]\n", true},
		{"H171", "1. Even if Alpha is red, then Beta is blue.\n2. Alpha is red.\n3. Therefore, Beta is blue. [1,2]\n", true},
		// H172/H173 were already rejected: the requested control classification
		// conflicts with only-if/unless, neither of which licenses forward MP.
		{"H172", "1. Only if Alpha is red, then Beta is blue.\n2. Alpha is red.\n3. Therefore, Beta is blue. [1,2]\n", false},
		{"H173", "1. Unless Alpha is red, then Beta is blue.\n2. Alpha is red.\n3. Therefore, Beta is blue. [1,2]\n", false},
		{"H174", "1. When Alpha is red, then Beta is blue.\n2. Alpha is red.\n3. Therefore, Beta is blue. [1,2]\n", true},
		{"H175", "1. Whenever Alpha is red, then Beta is blue.\n2. Alpha is red.\n3. Therefore, Beta is blue. [1,2]\n", true},
		{"H176", "1. The observer wonders aloud if Alpha is red.\n2. Alpha is red.\n3. Therefore, The observer wonders aloud. [1,2]\n", false},
		{"H177", "1. The observer asks the robot if Alpha is red.\n2. Alpha is red.\n3. Therefore, The observer asks the robot. [1,2]\n", false},
		{"H178", "1. The observer can tell if Alpha is red.\n2. Alpha is red.\n3. Therefore, The observer can tell. [1,2]\n", false},
		{"H179", "1. The observer is investigating if Alpha is red.\n2. Alpha is red.\n3. Therefore, The observer is investigating. [1,2]\n", false},
		{"H180", "1. The observer is checking if Alpha is red.\n2. Alpha is red.\n3. Therefore, The observer is checking. [1,2]\n", false},
		{"H181", "1. The observer has forgotten if Alpha is red.\n2. Alpha is red.\n3. Therefore, The observer has forgotten. [1,2]\n", false},
		{"H182", "1. The observer is debating if Alpha is red.\n2. Alpha is red.\n3. Therefore, The observer is debating. [1,2]\n", false},
		{"H183", "1. The observer is curious about if Alpha is red.\n2. Alpha is red.\n3. Therefore, The observer is curious about. [1,2]\n", false},
		{"H184", "1. The observer is uncertain about if Alpha is red.\n2. Alpha is red.\n3. Therefore, The observer is uncertain about. [1,2]\n", false},
		{"H185", "1. The observer is thinking about if Alpha is red.\n2. Alpha is red.\n3. Therefore, The observer is thinking about. [1,2]\n", false},
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			_, r := CheckEntailmentGate(tc.argument)
			if r.Passed != tc.valid {
				t.Fatalf("want valid=%v: %s", tc.valid, r.Message)
			}
			var id int
			fmt.Sscanf(tc.id, "H%d", &id)
			if id >= 97 && id <= 124 {
				sentence := strings.TrimSuffix(strings.SplitN(tc.argument, "\n", 2)[0][3:], ".")
				want := fmt.Sprintf("line 1 sentence %q: %s", sentence+".", operatorGroupingMessage)
				if r.Message != want {
					t.Fatalf("got %q, want %q", r.Message, want)
				}
			}
		})
	}
}

func TestAccuracyBindingClasses(t *testing.T) {
	for _, subject := range []string{"The observer hopes Alpha", "The observer suspects the large red robot", "It is probable that Alpha", "Maybe Alpha", "Apparently Alpha", "Supposedly Alpha", "Probably Alpha", "Sometimes Alpha"} {
		for _, mode := range []string{"schema", "witness", "UI"} {
			argument := "1. If X is red, then X is warm.\n2. " + subject + " is red.\n3. Therefore, " + subject + " is warm. [1,2]"
			if mode == "witness" {
				argument = "1. " + subject + " is red.\n2. Therefore, Something is red. [1]"
			}
			if mode == "UI" {
				argument = "1. Every red thing is warm.\n2. " + subject + " is a red thing.\n3. Therefore, " + subject + " is warm. [1,2]"
			}
			t.Run(mode+"/"+subject, func(t *testing.T) {
				_, r := CheckEntailmentGate(argument)
				if r.Passed {
					t.Fatal("embedded extraction accepted")
				}
			})
		}
	}
	for _, subject := range []string{"Each robot", "All robots", "Most robots", "Many robots", "Several robots", "More than two robots", "At least two robots", "The robot that is not blue"} {
		t.Run(subject, func(t *testing.T) {
			_, r := CheckEntailmentGate("1. If X is red, then X is warm.\n2. " + subject + " is red.\n3. Therefore, " + subject + " is warm. [1,2]")
			if !r.Passed {
				t.Fatal(r.Message)
			}
		})
	}
	for _, subject := range []string{"Exactly three robots", "Only Alpha", "All but two robots", "Almost all robots", "Nearly every robot", "At most three robots", "Fewer than three robots", "Few robots", "No robot", "None", "Neither robot"} {
		t.Run(subject, func(t *testing.T) {
			if schematicBindingAllowed(subject) {
				t.Fatal("non-upward binding accepted")
			}
		})
	}
}

func TestAccuracyMorphology(t *testing.T) {
	for _, pair := range [][2]string{{"clothes", "cloth"}, {"species", "specie"}, {"means", "mean"}, {"goods", "good"}, {"news", "new"}, {"odds", "odd"}, {"thanks", "thank"}, {"series", "serie"}, {"savings", "saving"}, {"earnings", "earning"}, {"belongings", "belonging"}, {"remains", "remain"}, {"surroundings", "surrounding"}, {"axes", "axe"}, {"axes", "axis"}, {"arms", "arm"}, {"customs", "custom"}, {"glasses", "glass"}} {
		t.Run(pair[0]+"/"+pair[1], func(t *testing.T) {
			if equalTerms(pair[0], pair[1]) {
				t.Fatal("distinct noun stems equated")
			}
		})
	}
	if !equalTerms("premises", "premise") {
		t.Fatal("regular plural rejected")
	}
}

func TestAccuracyNearbyControls(t *testing.T) {
	for _, tc := range []struct {
		name, argument string
		valid          bool
	}{
		{"wet if rain", "1. The ground is wet if it rains.\n2. It rains.\n3. Therefore, The ground is wet. [1,2]", true},
		{"dies if poison", "1. The philosopher dies if he drinks the hemlock.\n2. He drinks the hemlock.\n3. Therefore, The philosopher dies. [1,2]", true},
		{"atomic attitude MP", "1. If the light is on, then the observer believes Alpha is red.\n2. The light is on.\n3. Therefore, The observer believes Alpha is red. [1,2]", true},
		{"Oxford comma", "1. Alpha eats, shoots, and leaves.\n2. Therefore, Alpha eats, shoots and leaves. [1]", true},
		{"reported identity", "1. Alpha is identical to Beta.\n2. Alpha is reportedly red.\n3. Therefore, Beta is reportedly red. [1,2]", false},
		{"supposed identity", "1. Alpha is identical to Beta.\n2. Alpha is supposedly red.\n3. Therefore, Beta is supposedly red. [1,2]", false},
		{"known as identity", "1. Alpha is identical to Beta.\n2. Alpha is known as red.\n3. Therefore, Beta is known as red. [1,2]", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, r := CheckEntailmentGate(tc.argument)
			if r.Passed != tc.valid {
				t.Fatal(r.Message)
			}
		})
	}
	for _, s := range []string{"(Perhaps Alpha is red) and Beta is blue", "Perhaps (Alpha is red and Beta is blue)", "Both perhaps Alpha is red and Beta is blue", "The observer believes that both Alpha is red and Beta is blue"} {
		t.Run(s, func(t *testing.T) {
			if r := validateStatementGrouping(s, 1); !r.Passed {
				t.Fatal(r.Message)
			}
		})
	}
}
